package com.github.saxypandabear.songrequests.websocket.orchestrator

import com.github.saxypandabear.songrequests.websocket.TwitchSocket
import org.eclipse.jetty.websocket.client.WebSocketClient

/**
 * A ConnectionOrchestrator handles requests to connect to Twitch, and
 * requests to disconnect (unlisten), as well as load balancing between
 * different lower-level WebSocket clients in order to handle load.
 *
 * This should also handle the case where a WebSocket connection receives
 * a reconnect event from the server, in which case the entire WebSocket
 * connector would need to reconnect to the server, not just the individual
 * socket implementation.
 */
trait ConnectionOrchestrator {

  /**
   * Initiate a connection to Twitch with an internal WebSocket connection
   * @param channelId Twitch Channel ID to listen on
   * @param socketFactory Function that takes a channel ID and returns a
   *                      TwitchSocket
   */
  def connect(channelId: String, socketFactory: String => TwitchSocket): Unit

  /**
   * Stop listening to a connection to Twitch
   * @param channelId Twitch Channel ID to stop listening on
   */
  def disconnect(channelId: String): Unit

  /**
   * Reconnect/bounce the WebSocket client to force it to reconnect, because
   * a connector received a reconnect event from the server
   * @param channelId Twitch Channel ID that received a reconnect event
   */
  def reconnect(channelId: String): Unit

  /**
   * When an orchestrator is at capacity, the system should know to start
   * an auto-scaling event
   * @return true if the orchestrator is at capacity, false otherwise
   */
  def atCapacity: Boolean

  /**
   * Retrieve a snapshot map of the WebSocket clients and the channel IDs that
   * are connected to them.
   * @return a Map of WebSocket clients to the channel IDs that are connected
   *         to them
   */
  def connectionsToClients: Map[WebSocketClient, Set[String]]

  /**
   * Stop connections and perform any necessary clean-up
   */
  def stop(): Unit
  // TODO: determine what else is needed
}
