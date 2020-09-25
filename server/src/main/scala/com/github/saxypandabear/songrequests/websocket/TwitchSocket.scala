package com.github.saxypandabear.songrequests.websocket

import com.github.saxypandabear.songrequests.oauth.OauthTokenManager
import com.github.saxypandabear.songrequests.websocket.listener.WebSocketListener
import org.eclipse.jetty.websocket.api.Session
import org.eclipse.jetty.websocket.api.annotations._

// https://www.eclipse.org/jetty/documentation/current/jetty-websocket-client-api.html
/**
 * This class is the implementation that handles events from the WebSocket connection
 * TODO: For how this is defined, the socket is "stateless", but that may be subject to change
 * @param channelId         The Twitch channel ID that is associated with this connection
 * @param oauthTokenManager Manages the OAuth token necessary to authenticate against Twitch
 * @param listeners         List of listeners that act on each event from the WebSocket handler
 */
@WebSocket
class TwitchSocket(channelId: String,
                   oauthTokenManager: OauthTokenManager,
                   listeners: Seq[WebSocketListener] = Seq.empty) {

    /**
     * When we initially connect to the Twitch WebSocket, we should initiate a LISTEN
     * event to Twitch. See: https://dev.twitch.tv/docs/pubsub#connection-management
     * @param session Session state for this connection event
     */
    @OnWebSocketConnect
    def onConnect(session: Session): Unit = {
        this.listeners.foreach(_.onConnectEvent(channelId, session))
    }

    @OnWebSocketClose
    def onClose(session: Session, statusCode: Int, reason: String): Unit = {
        this.listeners.foreach(_.onCloseEvent(channelId, session, statusCode, reason))
    }

    @OnWebSocketError
    def onError(session: Session, cause: Throwable): Unit = {
        listeners.foreach(_.onErrorEvent(channelId, session, cause))
    }

    @OnWebSocketMessage
    def onMessage(session: Session, message: String): Unit = {
        listeners.foreach(_.onMessageEvent(channelId, session, message))
    }
}
