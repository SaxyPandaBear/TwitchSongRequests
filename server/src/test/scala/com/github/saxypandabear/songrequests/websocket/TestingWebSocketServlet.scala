package com.github.saxypandabear.songrequests.websocket

import javax.servlet.annotation.WebServlet
import org.eclipse.jetty.websocket.servlet.{WebSocketServlet, WebSocketServletFactory}

/**
 * This WebSocket servlet implementation uses a TestResponseSocket class to manage responding from the
 * server.
 */
@WebServlet(name = "Test WebSocket Server", urlPatterns = Array("/twitch"))
class TestingWebSocketServlet extends WebSocketServlet {
    override def configure(factory: WebSocketServletFactory): Unit = {
        factory.register(classOf[TestResponseSocket])
    }
}
