package com.github.saxypandabear.songrequests.websocket

import com.typesafe.scalalogging.StrictLogging
import org.eclipse.jetty.websocket.api.{Session, WebSocketAdapter}

class TestResponseSocket extends WebSocketAdapter with StrictLogging {
    override def onWebSocketConnect(sess: Session): Unit = {
        super.onWebSocketConnect(sess)
        logger.info("Received connect event")
        WebSocketTestingUtil.onConnect.release()
    }

    override def onWebSocketClose(statusCode: Int, reason: String): Unit = {
        super.onWebSocketClose(statusCode, reason)
        logger.info(s"Received close event: Status=$statusCode | Reason=$reason")
        WebSocketTestingUtil.onClose.release()
    }

    override def onWebSocketError(cause: Throwable): Unit = {
        super.onWebSocketError(cause)
        logger.error("Received error event", cause)
        WebSocketTestingUtil.onError.release()
    }

    override def onWebSocketText(message: String): Unit = {
        super.onWebSocketText(message)
        logger.info(s"Received message event: Message=$message")
        WebSocketTestingUtil.onMessage.release()
    }
}
