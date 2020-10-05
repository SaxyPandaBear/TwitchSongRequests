package com.github.saxypandabear.songrequests.websocket

import java.util.{Timer, TimerTask}

import com.github.saxypandabear.songrequests.oauth.OauthTokenManager
import com.github.saxypandabear.songrequests.websocket.listener.WebSocketListener
import org.eclipse.jetty.websocket.api.Session
import org.eclipse.jetty.websocket.api.annotations._

// https://www.eclipse.org/jetty/documentation/current/jetty-websocket-client-api.html
/**
 * This class is the implementation that handles events from the WebSocket connection
 * @param channelId         The Twitch channel ID that is associated with this connection
 * @param oauthTokenManager Manages the OAuth token necessary to authenticate against Twitch
 * @param listeners         List of listeners that act on each event from the WebSocket handler
 */
@WebSocket
class TwitchSocket(channelId: String,
                   oauthTokenManager: OauthTokenManager,
                   listeners: Seq[WebSocketListener] = Seq.empty,
                   pingFrequencyMs: Long = 60000) {

    private val spotifyUriPattern = "^(spotify:track:(\\w|\\d)+)$".r

    private var session: Session = _

    private val pingTask = new Timer()

    /**
     * When we initially connect to the Twitch WebSocket, we should initiate a LISTEN
     * event to Twitch. See: https://dev.twitch.tv/docs/pubsub#connection-management
     * @param session Session state for this connection event
     */
    @OnWebSocketConnect
    def onConnect(session: Session): Unit = {
        this.session = session
        this.listeners.foreach(_.onConnectEvent(channelId, session))
        startPing()
        sendListenEvent()
    }

    @OnWebSocketClose
    def onClose(statusCode: Int, reason: String): Unit = {
        this.listeners.foreach(_.onCloseEvent(channelId, session, statusCode, reason))

        stopPing()
    }

    @OnWebSocketError
    def onError(cause: Throwable): Unit = {
        listeners.foreach(_.onErrorEvent(channelId, session, cause))
    }

    @OnWebSocketMessage
    def onMessage(message: String): Unit = {
        listeners.foreach(_.onMessageEvent(channelId, session, message))
    }

    private[websocket] def inputMatchesSpotifyUri(userInput: String): Boolean = {
        userInput != null && spotifyUriPattern.findFirstIn(userInput.trim).isDefined
    }

    /**
     * Send a LISTEN event to the Twitch server
     */
    private def sendListenEvent(): Unit = {
        val listenMessage =
            s"""
               |{
               |  "type": "LISTEN",
               |  "nonce": "$channelId",
               |  "data": {
               |    "topics": ["channel-points-channel-v1.$channelId"],
               |    "auth_token": "${oauthTokenManager.getAccessToken}"
               |  }
               |}
               |""".stripMargin
        session.getRemote.sendString(listenMessage)
    }

    /**
     * We have to PING the server every X seconds in order to let Twitch know we are still listening
     */
    private def startPing(): Unit = {
        pingTask.scheduleAtFixedRate(new PingTimedTask(session), 0, pingFrequencyMs)
    }

    private def stopPing(): Unit = {
        pingTask.cancel()
    }
}

class PingTimedTask(session: Session) extends TimerTask {
    private final val PING_MESSAGE =
        """
          |{
          |  "type": "PING"
          |}
          |""".stripMargin

    override def run(): Unit = {
        session.getRemote.sendString(PING_MESSAGE)
    }
}
