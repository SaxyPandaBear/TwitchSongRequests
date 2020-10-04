package com.github.saxypandabear.songrequests.websocket

import java.net.URI
import java.util.UUID

import com.github.saxypandabear.songrequests.lib.{RotatingTestPort, UnitSpec}
import com.github.saxypandabear.songrequests.oauth.TestTokenManager
import com.github.saxypandabear.songrequests.websocket.listener.{LoggingWebSocketListener, TestingWebSocketListener}
import org.eclipse.jetty.server.Server
import org.eclipse.jetty.websocket.client.WebSocketClient
import org.scalatest.BeforeAndAfterEach
import org.scalatest.concurrent.Eventually
import org.scalatest.time.{Seconds, Span}

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

    // although we could test this along with another test, separating the assertion out for granularity
    "Connecting to a WebSocket server" should "send a LISTEN event to the server" ignore {

    }
}
