package com.github.saxypandabear.songrequests.websocket

import java.util.{Timer, TimerTask}

import com.github.saxypandabear.songrequests.oauth.OauthTokenManager
import com.github.saxypandabear.songrequests.queue.SongQueue
import com.github.saxypandabear.songrequests.util.JsonUtil.objectMapper
import com.github.saxypandabear.songrequests.websocket.listener.WebSocketListener
import org.eclipse.jetty.websocket.api.Session
import org.eclipse.jetty.websocket.api.annotations._

// https://www.eclipse.org/jetty/documentation/current/jetty-websocket-client-api.html
/**
 * This class is the implementation that handles events from the WebSocket connection
 * @param channelId         The Twitch channel ID that is associated with this connection
 * @param oauthTokenManager Manages the OAuth token necessary to authenticate against Twitch
 * @param songQueue         Queue that the socket submits parsed Spotify URIs to
 * @param listeners         List of listeners that act on each event from the WebSocket handler
 */
@WebSocket
class TwitchSocket(
    channelId: String,
    oauthTokenManager: OauthTokenManager,
    songQueue: SongQueue,
    listeners: Seq[WebSocketListener] = Seq.empty,
    pingFrequencyMs: Long = 60000
) {

  private val spotifyUriPattern = "^(spotify:track:(\\w|\\d)+)$".r
  private val pingTask          = new Timer()
  private var session: Session  = _

  /**
   * When we initially connect to the Twitch WebSocket, we should initiate a LISTEN
   * event to Twitch. See: https://dev.twitch.tv/docs/pubsub#connection-management
   * We should also begin a scheduled ping to the server to keep the connection alive
   * @param session Session state for this connection event
   */
  @OnWebSocketConnect
  def onConnect(session: Session): Unit = {
    this.session = session
    startPing()
    sendListenEvent()
    this.listeners.foreach(_.onConnectEvent(channelId, session))
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

  @OnWebSocketClose
  def onClose(statusCode: Int, reason: String): Unit = {
    stopPing()
    this.listeners.foreach(
        _.onCloseEvent(channelId, session, statusCode, reason)
    )
  }

  @OnWebSocketError
  def onError(cause: Throwable): Unit =
    listeners.foreach(_.onErrorEvent(channelId, session, cause))

  /**
   * When we receive a message from the server, there are a couple different types of messages that
   * should be expected, and each of them should be handled differently.
   * @param message String message from the server, expected to be a JSON string
   */
  @OnWebSocketMessage
  def onMessage(message: String): Unit = {
    val parsed       = objectMapper.readTree(message)
    val rootDataNode = parsed.get("data")
    listeners.foreach(_.onMessageEvent(channelId, session, message))
  }

  /**
   * We have to PING the server every X seconds in order to let Twitch know we are still listening
   */
  private def startPing(): Unit =
    pingTask.scheduleAtFixedRate(new PingTimedTask(session), 0, pingFrequencyMs)

  private def stopPing(): Unit = pingTask.cancel()

  private[websocket] def inputMatchesSpotifyUri(userInput: String): Boolean =
    userInput != null && spotifyUriPattern.findFirstIn(userInput.trim).isDefined
}

class PingTimedTask(session: Session) extends TimerTask {
  final private val PING_MESSAGE =
    """
          |{
          |  "type": "PING"
          |}
          |""".stripMargin

  override def run(): Unit =
    session.getRemote.sendStringByFuture(PING_MESSAGE)
}
