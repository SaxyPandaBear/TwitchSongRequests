package com.github.saxypandabear.songrequests.websocket.listener

import org.eclipse.jetty.websocket.api.Session

/**
 * Captures and stores state based on the events that the listener receives,
 * in internal TrieMaps so that we can inspect what events are triggered from
 * the WebSocket handler.
 */
class TestingWebSocketListener extends WebSocketListener {
    override def onConnectEvent(channelId: String, session: Session): Unit = {}

    override def onCloseEvent(channelId: String, session: Session, statusCode: Int, reason: String): Unit = {}

    override def onMessageEvent(channelId: String, session: Session, message: String): Unit = {}

    override def onErrorEvent(channelId: String, session: Session, error: Throwable): Unit = {}
}
