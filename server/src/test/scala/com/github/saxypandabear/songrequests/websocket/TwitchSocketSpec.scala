package com.github.saxypandabear.songrequests.websocket

import java.net.URI

import com.github.saxypandabear.songrequests.lib.{RotatingTestPort, UnitSpec}
import com.github.saxypandabear.songrequests.oauth.TestTokenManager
import com.github.saxypandabear.songrequests.websocket.listener.TestingWebSocketListener
import org.eclipse.jetty.server.Server
import org.eclipse.jetty.websocket.client.WebSocketClient
import org.scalatest.BeforeAndAfterEach

/**
 * Test class that should validate the functionality of the Twitch WebSocket handler.
 * It should test how the handler deals with connect requests, message events, errors, and close events.
 */
class TwitchSocketSpec extends UnitSpec
    with RotatingTestPort
    with BeforeAndAfterEach {

    private val testListener = new TestingWebSocketListener()
    private val testTokenManager = new TestTokenManager()

    private var server: Server = _
    private var port: Int = _

    override def beforeEach(): Unit = {
        testListener.flush()
        WebSocketTestingUtil.reset()

        port = randomPort()
        server = WebSocketTestingUtil.build(port)
        server.start()
    }

    override def afterEach(): Unit = {
        server.stop()
    }

    it should "work" in {
        val uri = new URI(s"ws://localhost:$port")
        val channelId = "abc123"

        val client = new WebSocketClient()
        val socket = new TwitchSocket(channelId, testTokenManager, Seq(testListener))
    }
}
