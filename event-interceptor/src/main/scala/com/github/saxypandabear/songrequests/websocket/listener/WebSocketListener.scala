package com.github.saxypandabear.songrequests.websocket.listener

import com.github.saxypandabear.songrequests.types.Types.ChannelId
import org.eclipse.jetty.websocket.api.Session

/**
 * A WebSocketListener does work when a WebSocket event is invoked. It expects a channel
 * ID because that helps to associate events to the specific channel it is listening on
 */
trait WebSocketListener {

  /**
   * Event handler that is executed when the socket connects to the server.
   * @param channelId Twitch channel ID
   * @param session WebSocket session
   */
  def onConnectEvent(channelId: ChannelId, session: Session): Unit

  /**
   * Event handler that is executed when the socket closes its connection to
   * the server
   * @param channelId Twitch channel ID
   * @param session WebSocket session
   * @param statusCode Status code for the disconnect
   * @param reason Reason for disconnecting
   */
  def onCloseEvent(
      channelId: ChannelId,
      session: Session,
      statusCode: Int,
      reason: String
  ): Unit

  /**
   * Event handler that is executed when the WebSocket client receives a
   * message from the server.
   * @param channelId Twitch channel ID
   * @param session WebSocket session
   * @param message Serialized message from the server
   */
  def onMessageEvent(
      channelId: ChannelId,
      session: Session,
      message: String
  ): Unit

  /**
   * Event handler that is executed when the WebSocket client encounters an
   * error.
   * @param channelId Twitch channel ID
   * @param session WebSocket session
   * @param error underlying WebSocket client exception
   */
  def onErrorEvent(
      channelId: ChannelId,
      session: Session,
      error: Throwable
  ): Unit
}
