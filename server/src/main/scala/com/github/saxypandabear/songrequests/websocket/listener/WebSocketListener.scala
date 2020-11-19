package com.github.saxypandabear.songrequests.websocket.listener

import org.eclipse.jetty.websocket.api.Session

/**
 * A WebSocketListener does work when a WebSocket event is invoked. It expects a channel
 * ID because that helps to associate events to the specific channel it is listening on
 */
trait WebSocketListener {
  def onConnectEvent(channelId: String, session: Session): Unit
  def onCloseEvent(
      channelId: String,
      session: Session,
      statusCode: Int,
      reason: String
  ): Unit
  def onMessageEvent(channelId: String, session: Session, message: String): Unit
  def onErrorEvent(channelId: String, session: Session, error: Throwable): Unit
}
