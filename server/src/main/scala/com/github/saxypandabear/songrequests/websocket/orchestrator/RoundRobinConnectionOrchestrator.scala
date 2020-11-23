package com.github.saxypandabear.songrequests.websocket.orchestrator

import java.net.URI
import java.util.concurrent.atomic.{AtomicBoolean, AtomicInteger}

import com.github.saxypandabear.songrequests.metric.CloudWatchMetricCollector
import com.github.saxypandabear.songrequests.websocket.TwitchSocket
import org.eclipse.jetty.websocket.client.WebSocketClient

import scala.collection.concurrent.TrieMap
import scala.collection.mutable

/**
 * A round robin implementation to how a connection orchestrator works.
 * It should check the health/capacity of each node before attempting
 * to add a connection to it.
 * This is required to be thread-safe because of how we expect messages to come
 * in from the API layer.
 * Scaling isn't a huge concern here, which is why we are okay with using
 * locks and synchronization on the internal set objects. Because there is a
 * hard limitation on how many clients we can have that connect from the same IP,
 * and how many connections are allowed per client, we should not have any
 * real scaling issues. We are dealing with, at most, hundreds of entities.
 * @param webSocketUri  URI that the WebSocket clients will connect to
 * @param maxNumSockets maximum number of sockets that this orchestrator
 *                      can use. this is constrained by the limit that
 *                      Twitch puts on how many client connections we
 *                      can have to their servers from a single IP address.
 *                      This is assumed to be an integer > 0
 * @param metrics       Class that handles collecting metrics to send to
 *                      CloudWatch
 */
class RoundRobinConnectionOrchestrator(
    metrics: CloudWatchMetricCollector,
    webSocketUri: URI,
    maxNumSockets: Int = 5,
    maxAllowedConnectionsPerClient: Int = 40
) extends ConnectionOrchestrator {

  // this handles the decision making of which WebSocket client to connect to
  private val position     = new AtomicInteger(0)
  private val isAtCapacity = new AtomicBoolean(false)

  // TODO: refactor this to also store a reference to the actual TwitchSocket
  //       object so that we can disconnect the socket at will.
  // associate a position to a tuple of a WebSocket client and the set of
  // channels connected to that WebSocket.
  // this was a deliberate design choice because I wanted to leverage the
  // benefits of using a TrieMap, but also wanted to ensure that the lookups
  // are tame. as such, it made more sense to make the key just the position,
  // an int, rather than making the key a composition of a position and the
  // WebSocket client, since the client wouldn't be helpful for the lookup.
  private[websocket] val indexedWebSocketConnections
      : TrieMap[Int, (WebSocketClient, mutable.HashSet[TwitchSocket])] =
    initInternalMap(
        maxNumSockets
    )

  /**
   * Initiate a connection to Twitch with an internal WebSocket connection.
   * Note: This only performs a connection when the orchestrator is not at
   *       capacity, and has a side effect of updating the isAtCapacity
   *       value
   * @param channelId     Twitch Channel ID to listen on
   * @param socketFactory Function that takes a channel ID and returns a Socket
   *                      implementation
   * @return true if successfully connected, false if at capacity
   */
  override def connect(
      channelId: String,
      socketFactory: String => TwitchSocket
  ): Boolean =
    if (!atCapacity) {
      val socket          = socketFactory(channelId)
      // get a valid WebSocket and connect
      var index           = position.getAndUpdate(p => rotate(p))
      var numTimesChecked = 0
      while (
          !(numTimesChecked < maxNumSockets) && !canClientAcceptNewConnection(
              index
          )
      ) {
        numTimesChecked += 1
        index = position.getAndUpdate(p => rotate(p))
      }
      if (canClientAcceptNewConnection(index)) {
        val (client, twitchSockets) = indexedWebSocketConnections(index)
        twitchSockets.synchronized {
          twitchSockets += socket
          client.connect(socket, webSocketUri).get() // connecting should be synchronous
        }
        true
      } else {
        // this means that we are at capacity and need to report as such.
        isAtCapacity.getAndSet(true)
        false
      }
    } else {
      false
    }

  /**
   * Stop listening to a connection to Twitch
   * @param channelId Twitch Channel ID to stop listening on
   */
  override def disconnect(channelId: String): Unit = {
    // 1. Find the socket with the channel ID

    // 2.
  }

  // TODO: implement me
  /**
   * Reconnect/bounce the WebSocket client to force it to reconnect, because
   * a connector received a reconnect event from the server
   * @param channelId Twitch Channel ID that received a reconnect event
   */
  override def reconnect(channelId: String): Unit = {}

  /**
   * When an orchestrator is at capacity, the system should know to start
   * an auto-scaling event
   * @return true if the orchestrator is at capacity, false otherwise
   */
  override def atCapacity: Boolean = isAtCapacity.get()

  /**
   * Stop connections and perform any necessary clean-up
   */
  override def stop(): Unit =
    for ((_, (client, _)) <- indexedWebSocketConnections)
      client.stop()

  /**
   * Retrieve a snapshot map of the WebSocket clients and the channel IDs that
   * are connected to them.
   * @return a Map of WebSocket clients to the channel IDs that are connected
   *         to them
   */
  override def connectionsToClients: Map[WebSocketClient, Set[String]] =
    indexedWebSocketConnections
      .readOnlySnapshot()
      .values
      .map { case (client, twitchSockets) =>
        client -> twitchSockets.map(_.channelId).toSet
      }
      .toMap

  /**
   * Increments the position and returns it. If position >= maxNumSockets,
   * then this wraps back to zero.
   * @param position current position
   * @return next valid position
   */
  private[orchestrator] def rotate(position: Int): Int = {
    val newPosition = position + 1
    val result      = if (newPosition >= maxNumSockets) 0 else newPosition
    result
  }

  private[orchestrator] def canClientAcceptNewConnection(
      position: Int
  ): Boolean =
    position < maxNumSockets && indexedWebSocketConnections
      .snapshot()
      .get(position)
      .exists { case (_, channelIds) =>
        channelIds.size < maxAllowedConnectionsPerClient
      }

  private def initInternalMap(
      numWebSocketClients: Int
  ): TrieMap[Int, (WebSocketClient, mutable.HashSet[TwitchSocket])] = {
    if (numWebSocketClients < 1) {
      throw new IllegalArgumentException(
          s"Orchestrator misconfigured - requires at least 1 client, but received $numWebSocketClients"
      )
    }
    val map =
      new TrieMap[Int, (WebSocketClient, mutable.HashSet[TwitchSocket])]()
    for (i <- 0 until numWebSocketClients) {
      // need to instantiate a WebSocket client object, and a new HashSet
      val twitchSockets = new mutable.HashSet[TwitchSocket]()
      val client        = new WebSocketClient()
      client.start()
      map.put(i, (client, twitchSockets))
    }
    map
  }
}
