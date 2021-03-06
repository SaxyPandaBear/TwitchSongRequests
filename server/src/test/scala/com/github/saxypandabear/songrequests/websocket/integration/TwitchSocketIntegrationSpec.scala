package com.github.saxypandabear.songrequests.websocket.integration

import java.net.URI
import java.util.UUID
import java.util.concurrent.Executors

import com.fasterxml.jackson.databind.JsonNode
import com.fasterxml.jackson.databind.node.ArrayNode
import com.github.saxypandabear.songrequests.lib.{
  DummyAmazonCloudWatch,
  RotatingTestPort,
  UnitSpec
}
import com.github.saxypandabear.songrequests.metric.CloudWatchMetricCollector
import com.github.saxypandabear.songrequests.oauth.TestTokenManager
import com.github.saxypandabear.songrequests.queue.InMemorySongQueue
import com.github.saxypandabear.songrequests.util.JsonUtil.objectMapper
import com.github.saxypandabear.songrequests.websocket.TwitchSocket
import com.github.saxypandabear.songrequests.websocket.lib.WebSocketTestingUtil
import com.github.saxypandabear.songrequests.websocket.listener.{
  LoggingWebSocketListener,
  TestingWebSocketListener
}
import org.eclipse.jetty.server.Server
import org.eclipse.jetty.websocket.client.WebSocketClient
import org.scalatest.concurrent.Eventually
import org.scalatest.tagobjects.Retryable
import org.scalatest.time.{Millis, Seconds, Span}
import org.scalatest.{BeforeAndAfterEach, Outcome, Retries}

import scala.collection.JavaConverters._
import scala.collection.mutable

/**
 * Test class that should validate the functionality of the Twitch WebSocket handler.
 * It should test how the handler deals with connect requests, message events, errors, and close events.
 */
class TwitchSocketIntegrationSpec
    extends UnitSpec
    with RotatingTestPort
    with BeforeAndAfterEach
    with Eventually
    with Retries {

  private val testListener                                = new TestingWebSocketListener()
  private val logListener                                 = new LoggingWebSocketListener()
  private val testingSongQueue                            = new InMemorySongQueue()
  private var testCloudWatchClient: DummyAmazonCloudWatch = _
  private var metricCollector: CloudWatchMetricCollector  = _
  private val executor                                    = Executors.newFixedThreadPool(1)

  // use this to assert after the server is shut down to make sure we properly
  // handle the disconnect events
  private val connectedChannelIds = new mutable.ArrayBuffer[String]()

  private var server: Server                   = _
  private var webSocketClient: WebSocketClient = _

  override def beforeEach(): Unit = {
    super.beforeEach()
    testCloudWatchClient = new DummyAmazonCloudWatch
    metricCollector =
      new CloudWatchMetricCollector(testCloudWatchClient, executor)

    connectedChannelIds.clear()
    testListener.clear()
    testingSongQueue.clear()
    WebSocketTestingUtil.reset()
    TestTokenManager.flush()

    server = WebSocketTestingUtil.build(port)
    server.start()
    webSocketClient = new WebSocketClient()
    webSocketClient.start()
  }

  // this also asserts
  override def afterEach(): Unit = {
    super.afterEach()

    WebSocketTestingUtil.onClose.acquire()
    webSocketClient.stop()
    // We expect to be able to proceed by this point
    eventually(timeout(Span(1, Seconds))) {
      WebSocketTestingUtil.onClose.acquire()
    }
    server.stop()
    for (channelId <- connectedChannelIds) {
      val closeEventOpt = testListener.closeEvents.get(channelId)
      closeEventOpt.isDefined should be(true)
      closeEventOpt.get.forall(event =>
        event == (1006, "Disconnected") || event == (1001, "Shutdown")
      ) should be(true)
    }
  }

  override def withFixture(test: NoArgTest): Outcome =
    if (isRetryable(test)) {
      withRetryOnFailure(super.withFixture(test))
    } else {
      super.withFixture(test)
    }

  // =================== Start onConnect Tests ===================
  /* Testing separate parts of the TwitchSocket onConnect for granularity */
  // TODO: This test can be flaky in CI, but always succeeds locally
  "Connecting to a WebSocket server" should "work" taggedAs Retryable in {
    val uri              = new URI(s"ws://localhost:$port")
    val channelId        = UUID.randomUUID().toString
    val testTokenManager = new TestTokenManager("abc123", "foo", "bar", "baz")

    val socket = new TwitchSocket(
        channelId,
        testTokenManager,
        testingSongQueue,
        metricCollector,
        Seq(testListener, logListener)
    )

    WebSocketTestingUtil.onConnect.acquire()
    webSocketClient.connect(socket, uri)
    WebSocketTestingUtil.onConnect.acquire()

    WebSocketTestingUtil.onConnect.availablePermits() should be(0)
    // wrap in an eventually block because the locking mechanism on the test
    // listener can cause a timing issue when running the CI workflow
    eventually(timeout(Span(100, Millis))) {
      testListener.connectEvents should contain theSameElementsAs Seq(channelId)
    }

    connectedChannelIds += channelId
  }

  "Connecting to a WebSocket server" should "send a LISTEN event to the server" in {
    val uri              = new URI(s"ws://localhost:$port")
    val channelId        = UUID.randomUUID().toString
    val clientId         = "abc123"
    val testTokenManager = new TestTokenManager(clientId, "foo", "bar", "baz")

    val socket = new TwitchSocket(
        channelId,
        testTokenManager,
        testingSongQueue,
        metricCollector,
        Seq(testListener, logListener)
    )

    WebSocketTestingUtil.onMessage.acquire()
    WebSocketTestingUtil.onConnect.acquire()
    webSocketClient.connect(socket, uri)
    WebSocketTestingUtil.onConnect.acquire()
    WebSocketTestingUtil.onMessage.acquire()

    WebSocketTestingUtil.onConnect.availablePermits() should be(0)
    WebSocketTestingUtil.onMessage.availablePermits() should be(0)

    WebSocketTestingUtil.listenMessages.length should be(1)
    TestTokenManager.clientIdsToTokens should have size 1
    testTokenManager.getAccessToken should be(
        TestTokenManager.clientIdsToTokens(clientId)
    )

    validateListenEvent(
        WebSocketTestingUtil.listenMessages.head,
        channelId,
        testTokenManager.getAccessToken
    )

    connectedChannelIds += channelId
  }

  "Connecting to a WebSocket server" should "start sending PING messages on a set frequency" in {
    val pingFrequencyMs  = 10
    val uri              = new URI(s"ws://localhost:$port")
    val channelId        = UUID.randomUUID().toString
    val testTokenManager = new TestTokenManager("abc123", "foo", "bar", "baz")

    val socket = new TwitchSocket(
        channelId,
        testTokenManager,
        testingSongQueue,
        metricCollector,
        Seq(testListener, logListener),
        pingFrequencyMs
    )

    WebSocketTestingUtil.onConnect.acquire()
    webSocketClient.connect(socket, uri)
    WebSocketTestingUtil.onConnect.acquire()

    WebSocketTestingUtil.onConnect.availablePermits() should be(0)

    // If we have a frequency of a ping every 10ms, we can expect roughly 10
    // pings in 100ms (more or less), but need a little wiggle room
    // each ping message should only contain a single field, that looks like:
    // { "type": "PING" }
    // We should also receive an equal amount of PONG replies from the server..
    // (with a little wiggle room because of timing)
    eventually(timeout(Span(150, Millis))) {
      val numPingMessages = WebSocketTestingUtil.pingMessages.length
      numPingMessages should be >= 10
      WebSocketTestingUtil.pingMessages.forall { pingMessage =>
        pingMessage.has("type") &&
        pingMessage.get("type").asText() == "PING" &&
        pingMessage.fields().asScala.length == 1
      } should be(true)
      testListener.messageEvents
        .getOrElse(channelId, fail(s"Channel ID $channelId does not exist"))
        .map(objectMapper.readTree)
        .count(node =>
          node.has("type") && node.get("type").asText() == "PONG"
        ) should
        be(numPingMessages +- 2)
    }

    connectedChannelIds += channelId
  }
  // =================== End onConnect Tests ===================

  // =================== Start onMessage Tests ===================
  /* The main piece to test with onMessage is how it parses and handles input
   * from the server. */
  "Receiving a redemption event from the server" should "attempt to queue a song" in {
    val uri              = new URI(s"ws://localhost:$port")
    val channelId        = UUID.randomUUID().toString
    val clientId         = "abc123"
    val testTokenManager = new TestTokenManager(clientId, "foo", "bar", "baz")

    val socket = new TwitchSocket(
        channelId,
        testTokenManager,
        testingSongQueue,
        metricCollector,
        Seq(testListener, logListener)
    )

    WebSocketTestingUtil.onMessage.acquire()
    WebSocketTestingUtil.onConnect.acquire()
    webSocketClient.connect(socket, uri)
    WebSocketTestingUtil.onMessage.acquire()
    WebSocketTestingUtil.onConnect.acquire()

    WebSocketTestingUtil.startSending.acquire()
    WebSocketTestingUtil.initializeSemaphoresForSending(
        numRedeem = 1,
        shouldSendRedeem = true,
        numReconnect = 0,
        shouldSendReconnect = false
    )
    WebSocketTestingUtil.startSending.acquire() // wait until the server starts sending messages

    // now that the timer will send 1 message, we can assert against it.
    eventually(timeout(Span(100, Millis))) {
      // there should only be one message that matters (non PONG), but the
      // listener captures all of the messages.
      testListener.messageEvents
        .getOrElse(channelId, fail(s"Channel ID $channelId does not exist"))
        .map(objectMapper.readTree)
        .count(node =>
          node.has("type") && node.get("type").asText() != "PONG"
        ) should be(1)
      testingSongQueue.queued.getOrElse(
          channelId,
          fail(s"Channel ID $channelId does not exist")
      ) should contain theSameElementsAs WebSocketTestingUtil.spotifyUris
    }
  }

  "Receiving multiple redemption events" should "queue them all separately" in {
    val uri              = new URI(s"ws://localhost:$port")
    val channelId        = UUID.randomUUID().toString
    val clientId         = "abc123"
    val testTokenManager = new TestTokenManager(clientId, "foo", "bar", "baz")
    val numMessages      = 5

    val socket = new TwitchSocket(
        channelId,
        testTokenManager,
        testingSongQueue,
        metricCollector,
        Seq(testListener, logListener)
    )

    WebSocketTestingUtil.onMessage.acquire()
    WebSocketTestingUtil.onConnect.acquire()
    webSocketClient.connect(socket, uri)
    WebSocketTestingUtil.onMessage.acquire()
    WebSocketTestingUtil.onConnect.acquire()

    WebSocketTestingUtil.startSending.acquire()
    WebSocketTestingUtil.initializeSemaphoresForSending(
        numRedeem = numMessages,
        shouldSendRedeem = true,
        numReconnect = 0,
        shouldSendReconnect = false
    )
    WebSocketTestingUtil.startSending.acquire() // wait until the server starts sending messages

    eventually(timeout(Span(150, Millis))) {
      WebSocketTestingUtil.doneSending.get() should be(true)
      testListener.messageEvents
        .getOrElse(channelId, fail(s"Channel ID $channelId does not exist"))
        .map(objectMapper.readTree)
        .count(node =>
          node.has("type") && node.get("type").asText() != "PONG"
        ) should be(numMessages)
      testingSongQueue.queued.getOrElse(
          channelId,
          fail(s"Channel ID $channelId does not exist")
      ) should contain theSameElementsAs WebSocketTestingUtil.spotifyUris
    }
  }

  // TODO: may need to hold off on this testing until the load balancer is
  //       implemented, since it requires performing a reconnect of the
  //       web socket. Leaving this test ignored in the meantime
  "Receiving a reconnect message" should "trigger a new LISTEN message" ignore {
    val uri              = new URI(s"ws://localhost:$port")
    val channelId        = UUID.randomUUID().toString
    val clientId         = "abc123"
    val testTokenManager = new TestTokenManager(clientId, "foo", "bar", "baz")

    val socket = new TwitchSocket(
        channelId,
        testTokenManager,
        testingSongQueue,
        metricCollector,
        Seq(testListener, logListener)
    )

    WebSocketTestingUtil.onMessage.acquire()
    WebSocketTestingUtil.onConnect.acquire()
    webSocketClient.connect(socket, uri)
    WebSocketTestingUtil.onMessage.acquire()
    WebSocketTestingUtil.onConnect.acquire()

    val numListenEvents = WebSocketTestingUtil.listenMessages.length
    numListenEvents should be(1)

    WebSocketTestingUtil.startSending.acquire()
    WebSocketTestingUtil.initializeSemaphoresForSending(
        numRedeem = 0,
        shouldSendRedeem = false,
        numReconnect = 1,
        shouldSendReconnect = true
    )
    WebSocketTestingUtil.startSending.acquire() // wait until the server starts sending messages

    eventually(timeout(Span(50, Millis))) {
      WebSocketTestingUtil.doneSending.get() should be(true)
    }
  }

  // =================== End onMessage Tests ===================

  private def validateListenEvent(
      event: JsonNode,
      channelId: String,
      oauthToken: String
  ): Unit = {
    event.has("nonce") should be(true)
    event.get("nonce").asText() should be(channelId)

    val dataNode = event.get("data")

    val topicsNode = dataNode.get("topics")
    topicsNode.isArray should be(true)
    topicsNode.asInstanceOf[ArrayNode].get(0).asText() should be(
        s"channel-points-channel-v1.$channelId"
    )

    dataNode.get("auth_token").asText() should be(oauthToken)
  }
}
