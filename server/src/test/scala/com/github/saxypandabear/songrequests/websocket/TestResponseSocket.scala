package com.github.saxypandabear.songrequests.websocket

import com.fasterxml.jackson.databind.JsonNode
import com.github.saxypandabear.songrequests.util.JsonUtil.objectMapper
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

    // note: do not release the semaphore on a ping event
    override def onWebSocketText(message: String): Unit = {
        super.onWebSocketText(message)
        logger.info(s"Received message event: Message=$message")

        // need to acknowledge PING, LISTEN, and UNLISTEN type messages
        val parsedJson = objectMapper.readTree(message)
        handleMessage(parsedJson)
    }

    private def handleMessage(jsonMessage: JsonNode): Unit = {
        val messageType = jsonMessage.get("type").asText()
        logger.info("Message type received: {}", messageType)
        messageType match {
            case "PING" => handlePingMessage(jsonMessage)
            case "LISTEN" => handleListenMessage(jsonMessage)
            case "UNLISTEN" => handleUnlistenMessage(jsonMessage)
        }
    }

    private def handlePingMessage(jsonNode: JsonNode): Unit = {
        logger.info("Ping message received")
        WebSocketTestingUtil.pingMessages += jsonNode
    }

    private def handleListenMessage(jsonNode: JsonNode): Unit = {
        logger.info("Listen message received")
        WebSocketTestingUtil.listenMessages += jsonNode
        WebSocketTestingUtil.onMessage.release()
    }

    private def handleUnlistenMessage(jsonNode: JsonNode): Unit = {
        logger.info("Unlisten message received")
        WebSocketTestingUtil.unlistenMessages += jsonNode
        WebSocketTestingUtil.onMessage.release()
    }
}
