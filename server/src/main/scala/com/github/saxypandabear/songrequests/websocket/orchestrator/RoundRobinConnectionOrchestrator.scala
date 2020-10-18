package com.github.saxypandabear.songrequests.websocket.orchestrator

import java.util.concurrent.atomic.AtomicInteger

import com.github.saxypandabear.songrequests.ddb.ConnectionDataStore
import com.github.saxypandabear.songrequests.oauth.OauthTokenManagerFactory
import com.github.saxypandabear.songrequests.queue.SongQueue
import com.github.saxypandabear.songrequests.websocket.listener.WebSocketListener
import org.eclipse.jetty.websocket.client.WebSocketClient

import scala.collection.concurrent.TrieMap
import scala.collection.mutable

// TODO: implement the stuff in here
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
 * @param clientId            client id of the application
 * @param clientSecret        client secret of the application
 * @param refreshUri          URI for the OAuth token managers to re-authenticate
 *                            to Twitch
 * @param dataStore           data store that contains all of the connection
 *                            details
 * @param songQueue           queue that interfaces with Spotify
 * @param tokenManagerFactory implementation for how this orchestrator will
 *                            create token managers. this gets fed to the
 *                            TwitchSocketFactory, but we accept anything
 *                            that extends from the factory trait here, since
 *                            that gives us more leeway during tests.
 * @param maxNumSockets       maximum number of sockets that this orchestrator
 *                            can use. this is constrained by the limit that
 *                            Twitch puts on how many client connections we
 *                            can have to their servers from a single IP address.
 *                            This is assumed to be an integer > 0
 * @param webSocketListeners  list of web socket listeners to pass in when
 *                            constructing the socket implementation
 * @param pingFrequencyMs     configurable frequency for when the WebSocket
 *                            client will ping the server
 */
class RoundRobinConnectionOrchestrator(
    clientId: String,
    clientSecret: String,
    refreshUri: String,
    dataStore: ConnectionDataStore,
    songQueue: SongQueue,
    tokenManagerFactory: OauthTokenManagerFactory,
    maxNumSockets: Int = 5,
    webSocketListeners: Seq[WebSocketListener] = Seq.empty,
    pingFrequencyMs: Long = 60000
) extends ConnectionOrchestrator {

  private val MAX_ALLOWED_CONNECTIONS_PER_CLIENT = 40

  // this handles the decision making of which WebSocket client to connect to
  private val position = new AtomicInteger(0)

  // associate a position to a tuple of a WebSocket client and the set of
  // channels connected to that WebSocket.
  // this was a deliberate design choice because I wanted to leverage the
  // benefits of using a TrieMap, but also wanted to ensure that the lookups
  // are tame. as such, it made more sense to make the key just the position,
  // an int, rather than making the key a composition of a position and the
  // WebSocket client, since the client wouldn't be helpful for the lookup.
  private val indexedWebSocketConnections
      : TrieMap[Int, (WebSocketClient, mutable.HashSet[String])] =
    initInternalMap(
        maxNumSockets
    )

  /**
   * Initiate a connection to Twitch with an internal WebSocket connection
   * @param channelId Twitch Channel ID to listen on
   */
  override def connect(channelId: String): Unit = {}

  /**
   * Stop listening to a connection to Twitch
   * @param channelId Twitch Channel ID to stop listening on
   */
  override def disconnect(channelId: String): Unit = ???

  /**
   * Reconnect/bounce the WebSocket client to force it to reconnect, because
   * a connector received a reconnect event from the server
   * @param channelId Twitch Channel ID that received a reconnect event
   */
  override def reconnect(channelId: String): Unit = ???

  /**
   * When an orchestrator is at capacity, the system should know to start
   * an auto-scaling event
   * @return true if the orchestrator is at capacity, false otherwise
   */
  override def atCapacity: Boolean = ???

  /**
   * Stop connections and perform any necessary clean-up
   */
  override def stop(): Unit = {}

  /**
   * Synchronous, blocking method that performs a get and set operation
   * on the AtomicInteger to get and update the position that the orchestrator
   * will use for the next invocation to connect. Note that this also needs
   * to check how many connections there currently are for a given WebSocket
   * client. This is not a simple ++ operation. This also wraps around, so
   * the return value will always be a valid, positive integer between
   * [0, maxNumSockets).
   * If there is no available
   * @return the current valid index to use for a connection, or -1 if there
   *         is no longer capacity to handle more connections
   */
  private def getAndRotate(): Int =
    position.synchronized {
      val currentPosition = position.get()
      var nextPosition    = currentPosition + 1
      if (nextPosition == maxNumSockets) {
        nextPosition = 0 // reset
      }
      // before we can proceed, we need to ensure that we have not overloaded
      // an existing WebSocket client. Check how many clients are already
      // connected to the nextPosition client. If it is above the allowed
      // threshold, then we have to increment and check again

      currentPosition
    }

  private[orchestrator] def canClientAcceptNewConnection(
      position: Int
  ): Boolean =
    indexedWebSocketConnections.snapshot().get(position).exists {
      case (_, channelIds) =>
        channelIds.size < MAX_ALLOWED_CONNECTIONS_PER_CLIENT
    }

  private def initInternalMap(
      numWebSocketClients: Int
  ): TrieMap[Int, (WebSocketClient, mutable.HashSet[String])] = {
    if (numWebSocketClients < 1) {
      throw new IllegalArgumentException(
          s"Orchestrator misconfigured - requires at least 1 client, but received $numWebSocketClients"
      )
    }
    val map = new TrieMap[Int, (WebSocketClient, mutable.HashSet[String])]()
    for (i <- 0 until numWebSocketClients) {
      // need to instantiate a WebSocket client object, and a new HashSet
      val channelIds = new mutable.HashSet[String]()
      val client     = new WebSocketClient()
      client.start()
      map.put(i, (client, channelIds))
    }
    map
  }
}
