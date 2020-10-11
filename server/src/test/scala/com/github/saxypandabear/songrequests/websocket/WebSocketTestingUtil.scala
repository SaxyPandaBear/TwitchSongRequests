package com.github.saxypandabear.songrequests.websocket

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
  }

  def createRedeemEvent(): String =
    s"""
       |{
       |  "type": "MESSAGE",
       |  "data": {
       |    "type": "reward-redeemed",
       |    "data": {
       |       "timestamp": "2019-11-12T01:29:34.98329743Z",
       |       "redemption": {
       |          "id": "9203c6f0-51b6-4d1d-a9ae-8eafdb0d6d47",
       |          "user": {
       |            "id": "30515034",
       |            "login": "davethecust",
       |            "display_name": "davethecust"
       |          },
       |        "channel_id": "30515034",
       |        "redeemed_at": "2019-12-11T18:52:53.128421623Z",
       |        "reward": {
       |          "id": "6ef17bb2-e5ae-432e-8b3f-5ac4dd774668",
       |          "channel_id": "30515034",
       |          "title": "A Song Request",
       |          "prompt": "Some request \n",
       |          "cost": 10,
       |          "is_user_input_required": true,
       |          "is_sub_only": false,
       |          "image": {
       |            "url_1x": "https://static-cdn.jtvnw.net/custom-reward-images/30515034/6ef17bb2-e5ae-432e-8b3f-5ac4dd774668/7bcd9ca8-da17-42c9-800a-2f08832e5d4b/custom-1.png",
       |            "url_2x": "https://static-cdn.jtvnw.net/custom-reward-images/30515034/6ef17bb2-e5ae-432e-8b3f-5ac4dd774668/7bcd9ca8-da17-42c9-800a-2f08832e5d4b/custom-2.png",
       |            "url_4x": "https://static-cdn.jtvnw.net/custom-reward-images/30515034/6ef17bb2-e5ae-432e-8b3f-5ac4dd774668/7bcd9ca8-da17-42c9-800a-2f08832e5d4b/custom-4.png"
       |          },
       |          "default_image": {
       |            "url_1x": "https://static-cdn.jtvnw.net/custom-reward-images/default-1.png",
       |            "url_2x": "https://static-cdn.jtvnw.net/custom-reward-images/default-2.png",
       |            "url_4x": "https://static-cdn.jtvnw.net/custom-reward-images/default-4.png"
       |          },
       |          "background_color": "#00C7AC",
       |          "is_enabled": true,
       |          "is_paused": false,
       |          "is_in_stock": true,
       |          "max_per_stream": { "is_enabled": false, "max_per_stream": 0 },
       |          "should_redemptions_skip_request_queue": true
       |        },
       |        "user_input": "${generateSpotifyUri()}",
       |        "status": "FULFILLED"
       |      }
       |    }
       |  }
       |}
       |""".stripMargin

  def createReconnectEvent(): String =
    """
      |{
      |  "type": "MESSAGE",
      |  "data": 
      |}
      |""".stripMargin

  private def generateSpotifyUri(): String =
    s"spotify:track:${UUID.randomUUID().toString}"
}
