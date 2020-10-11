package com.github.saxypandabear.songrequests.websocket

import java.net.URI
import java.util.UUID

import com.fasterxml.jackson.databind.JsonNode
import com.fasterxml.jackson.databind.node.ArrayNode
import com.github.saxypandabear.songrequests.lib.{RotatingTestPort, UnitSpec}
import com.github.saxypandabear.songrequests.oauth.TestTokenManager
import com.github.saxypandabear.songrequests.websocket.listener.{LoggingWebSocketListener, TestingWebSocketListener}
import org.eclipse.jetty.server.Server
import org.eclipse.jetty.websocket.client.WebSocketClient
import org.scalatest.BeforeAndAfterEach
import org.scalatest.concurrent.Eventually
import org.scalatest.time.{Millis, Seconds, Span}

import scala.collection.JavaConverters._
import scala.collection.mutable

/**
 * Test class that should validate the functionality of the Twitch WebSocket handler.
 * It should test how the handler deals with connect requests, message events, errors, and close events.
 */
class TwitchSocketIntegrationSpec extends UnitSpec
    with RotatingTestPort
    with BeforeAndAfterEach
    with Eventually {

    private val testListener = new TestingWebSocketListener()
    private val logListener = new LoggingWebSocketListener()

    // use this to assert after the server is shut down to make sure we properly handle the disconnect events
    private val connectedChannelIds = new mutable.ArrayBuffer[String]()

    private var server: Server = _
    private var webSocketClient: WebSocketClient = _

    override def beforeEach(): Unit = {
        super.beforeEach()

        connectedChannelIds.clear()
        testListener.flush()
        WebSocketTestingUtil.reset()

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
        eventually(timeout(Span(3, Seconds))) {
            WebSocketTestingUtil.onClose.acquire()
        }
        server.stop()
        for (channelId <- connectedChannelIds) {
            val closeEventOpt = testListener.closeEvents.get(channelId)
            closeEventOpt.isDefined should be(true)
            closeEventOpt.get should be((1006, "Disconnected"))
        }
    }

    // =================== Start onConnect Tests ===================
    /*
        Testing separate parts of the TwitchSocket onConnect for granularity
    */
    "Connecting to a WebSocket server" should "work" in {
        val uri = new URI(s"ws://localhost:$port")
        val channelId = UUID.randomUUID().toString
        val testTokenManager = new TestTokenManager("abc123", "foo", "bar", "baz")

        val socket = new TwitchSocket(channelId, testTokenManager, Seq(testListener, logListener))

        WebSocketTestingUtil.onConnect.acquire()
        webSocketClient.connect(socket, uri)
        WebSocketTestingUtil.onConnect.acquire()

        WebSocketTestingUtil.onConnect.availablePermits() should be(0)
        testListener.connectEvents should contain theSameElementsAs Seq(channelId)

        connectedChannelIds += channelId
    }

    "Connecting to a WebSocket server" should "send a LISTEN event to the server" in {
        val uri = new URI(s"ws://localhost:$port")
        val channelId = UUID.randomUUID().toString
        val clientId = "abc123"
        val testTokenManager = new TestTokenManager(clientId, "foo", "bar", "baz")

        val socket = new TwitchSocket(channelId, testTokenManager, Seq(testListener, logListener))

        WebSocketTestingUtil.onMessage.acquire()
        WebSocketTestingUtil.onConnect.acquire()
        webSocketClient.connect(socket, uri)
        WebSocketTestingUtil.onConnect.acquire()
        WebSocketTestingUtil.onMessage.acquire()

        WebSocketTestingUtil.onConnect.availablePermits() should be(0)
        WebSocketTestingUtil.onMessage.availablePermits() should be(0)

        WebSocketTestingUtil.listenMessages.length should be(1)
        TestTokenManager.clientIdsToTokens should have size 1
        testTokenManager.getAccessToken should be(TestTokenManager.clientIdsToTokens(clientId))

        validateListenEvent(WebSocketTestingUtil.listenMessages.head, channelId, testTokenManager.getAccessToken)

        connectedChannelIds += channelId
    }

    "Connecting to a WebSocket server" should "start sending PING messages on a set frequency" in {
        val pingFrequencyMs = 10
        val uri = new URI(s"ws://localhost:$port")
        val channelId = UUID.randomUUID().toString
        val testTokenManager = new TestTokenManager("abc123", "foo", "bar", "baz")

        val socket = new TwitchSocket(channelId, testTokenManager, Seq(testListener, logListener), pingFrequencyMs)

        WebSocketTestingUtil.onConnect.acquire()
        webSocketClient.connect(socket, uri)
        WebSocketTestingUtil.onConnect.acquire()

        WebSocketTestingUtil.onConnect.availablePermits() should be(0)

        // If we have a frequency of a ping every 10ms, we can expect roughly 10 pings in 100ms (more or less)
        // each ping message should only contain a single field, that looks like:
        // { "type": "PING" }
        eventually(timeout(Span(100, Millis))) {
            WebSocketTestingUtil.pingMessages.length should be(10 +- 1)
            WebSocketTestingUtil.pingMessages.forall(pingMessage => {
                pingMessage.has("type") &&
                    pingMessage.get("type").asText() == "PING" &&
                    pingMessage.fields().asScala.length == 1
            }) should be(true)
        }

        connectedChannelIds += channelId
    }

    // =================== End onConnect Tests ===================

    private def validateListenEvent(event: JsonNode, channelId: String, oauthToken: String): Unit = {
        event.has("nonce") should be(true)
        event.get("nonce").asText() should be(channelId)

        val dataNode = event.get("data")

        val topicsNode = dataNode.get("topics")
        topicsNode.isArray should be(true)
        topicsNode.asInstanceOf[ArrayNode].get(0).asText() should be(s"channel-points-channel-v1.$channelId")

        dataNode.get("auth_token").asText() should be(oauthToken)
    }
}
