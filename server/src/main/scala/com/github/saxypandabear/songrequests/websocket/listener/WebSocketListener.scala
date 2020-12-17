package com.github.saxypandabear.songrequests.websocket.listener

import com.github.saxypandabear.songrequests.types.Types.ChannelId
import org.eclipse.jetty.websocket.api.Session

/**
 * A WebSocketListener does work when a WebSocket event is invoked. It expects a channel
 * ID because that helps to associate events to the specific channel it is listening on
 */
trait WebSocketListener {
  def onConnectEvent(channelId: ChannelId, session: Session): Unit
  def onCloseEvent(
      channelId: ChannelId,
      session: Session,
      statusCode: Int,
      reason: String
  ): Unit
  def onMessageEvent(
      channelId: ChannelId,
      session: Session,
      message: String
  ): Unit
  def onErrorEvent(
      channelId: ChannelId,
      session: Session,
      error: Throwable
  ): Unit
}
