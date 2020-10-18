package com.github.saxypandabear.songrequests.websocket.lib

import java.util.UUID
import java.util.concurrent.Semaphore
import java.util.concurrent.atomic.AtomicBoolean

import com.fasterxml.jackson.databind.JsonNode
import org.eclipse.jetty.server.Server
import org.eclipse.jetty.servlet.ServletContextHandler

import scala.collection.mutable

object WebSocketTestingUtil {

  // keeps track of the "message types" that the Twitch Socket can send to the
  // server
  val acceptedMessageTypes     = Set("PING", "LISTEN", "UNLISTEN")
  // keeps track of the channel IDs that are allowed to interact with the
  // server.
  // this helps to manage paths for testing, i.e.: which channel IDs trigger
  // error events, etc.
  val acceptedChannelIds       = Set("abc123")
  // Stores server state that tracks which events occur when handling messages
  val pingMessages             = new mutable.ArrayBuffer[JsonNode]()
  val listenMessages           = new mutable.ArrayBuffer[JsonNode]()
  val unlistenMessages         = new mutable.ArrayBuffer[JsonNode]()
  // Locking mechanisms to block on events
  var onConnect                = new Semaphore(1)
  var onClose                  = new Semaphore(1)
  var onError                  = new Semaphore(1)
  var onMessage                = new Semaphore(1)
  // so long as there are permits on the below semaphores, the server will
  // send the specified messages at a set frequency
  var shouldSendRedeemEvent    = new AtomicBoolean(false)
  var redeemEvents             = new Semaphore(1)
  var shouldSendReconnectEvent = new AtomicBoolean(false)
  var reconnectEvents          = new Semaphore(1)
  val sendFrequencyMs          = 10
  var startSending             = new Semaphore(1)
  val doneSending              = new AtomicBoolean(false)
  val spotifyUris              = new mutable.ArrayBuffer[String]()

  /**
   * Use this method to help with setting up the test server to send data
   * to the client at a set frequency
   * @param numRedeem           number of redeem events to send
   * @param shouldSendRedeem    should the server send redeem messages
   * @param numReconnect        number of reconnect events to send
   * @param shouldSendReconnect should the server send reconnect messages
   */
  def initializeSemaphoresForSending(
      numRedeem: Int,
      shouldSendRedeem: Boolean,
      numReconnect: Int,
      shouldSendReconnect: Boolean
  ): Unit = {
    redeemEvents = new Semaphore(numRedeem)
    reconnectEvents = new Semaphore(numReconnect)
    shouldSendRedeemEvent.getAndSet(shouldSendRedeem)
    shouldSendReconnectEvent.getAndSet(shouldSendReconnect)
  }

  def build(port: Int): Server = {
    val server = new Server(port)
    server.setStopAtShutdown(true)
    server.setStopTimeout(0)

    val ctx = new ServletContextHandler(ServletContextHandler.NO_SESSIONS)
    ctx.setContextPath("/")

    ctx.addServlet(classOf[TestingWebSocketServlet], "/")

    server.setHandler(ctx)
    server
  }

  /**
   * Resets the semaphores used in testing
   */
  def reset(): Unit = {
    onConnect = new Semaphore(1)
    onClose = new Semaphore(1)
    onError = new Semaphore(1)
    onMessage = new Semaphore(1)

    pingMessages.clear()
    listenMessages.clear()
    unlistenMessages.clear()

    shouldSendRedeemEvent = new AtomicBoolean(false)
    redeemEvents = new Semaphore(1)
    shouldSendReconnectEvent = new AtomicBoolean(false)
    reconnectEvents = new Semaphore(1)
    startSending = new Semaphore(1)

    spotifyUris.clear()
  }

  //noinspection ScalaStyle
  // note: this is super long because the nested message is a string
  // representation of a JSON message.
  def createRedeemEvent(): String =
    s"""{
       |    "type": "MESSAGE",
       |    "data": {
       |        "topic": "channel-points-channel-v1.106060203",
       |        "message": "{\\"type\\":\\"reward-redeemed\\",\\"data\\":{\\"timestamp\\":\\"2020-08-23T20:21:56.588735036Z\\",\\"redemption\\":{\\"id\\":\\"897dd20c-ec7f-42da-9e0a-610091785a4d\\",\\"user\\":{\\"id\\":\\"106060203\\",\\"login\\":\\"saxypandabear\\",\\"display_name\\":\\"SaxyPandaBear\\"},\\"channel_id\\":\\"106060203\\",\\"redeemed_at\\":\\"2020-08-23T20:21:56.588735036Z\\",\\"reward\\":{\\"id\\":\\"ca20aaa2-5fa8-4b29-a9a6-34275ee911f4\\",\\"channel_id\\":\\"106060203\\",\\"title\\":\\"Song Request\\",\\"prompt\\":\\"Only applies for music streams. Request a song you want me to attempt to learn by ear.\\",\\"cost\\":10000,\\"is_user_input_required\\":true,\\"is_sub_only\\":false,\\"image\\":null,\\"default_image\\":{\\"url_1x\\":\\"https://static-cdn.jtvnw.net/custom-reward-images/default-1.png\\",\\"url_2x\\":\\"https://static-cdn.jtvnw.net/custom-reward-images/default-2.png\\",\\"url_4x\\":\\"https://static-cdn.jtvnw.net/custom-reward-images/default-4.png\\"},\\"background_color\\":\\"#FA2929\\",\\"is_enabled\\":true,\\"is_paused\\":false,\\"is_in_stock\\":true,\\"max_per_stream\\":{\\"is_enabled\\":false,\\"max_per_stream\\":0},\\"should_redemptions_skip_request_queue\\":false,\\"template_id\\":null,\\"updated_for_indicator_at\\":\\"2020-01-01T15:11:26.647212555Z\\",\\"max_per_user_per_stream\\":{\\"is_enabled\\":false,\\"max_per_user_per_stream\\":0},\\"global_cooldown\\":{\\"is_enabled\\":false,\\"global_cooldown_seconds\\":0},\\"redemptions_redeemed_current_stream\\":0,\\"cooldown_expires_at\\":null},\\"user_input\\":\\"${generateSpotifyUri()}\\",\\"status\\":\\"UNFULFILLED\\"}}}"
       |    }
       |}
       |""".stripMargin

  // TODO: figure out which level has the RECONNECT type
  //       (is it at the root?)
  def createReconnectEvent(): String =
    """
      |{
      |  "type": "RECONNECT"
      |}
      |""".stripMargin

  private def generateSpotifyUri(): String = {
    val uri = s"spotify:track:${UUID.randomUUID().toString.replace("-", "")}"
    spotifyUris += uri
    uri
  }
}
