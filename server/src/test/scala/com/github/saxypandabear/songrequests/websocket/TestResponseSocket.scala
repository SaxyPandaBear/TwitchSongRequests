package com.github.saxypandabear.songrequests.websocket

import java.util.{Timer, TimerTask}

import com.fasterxml.jackson.databind.JsonNode
import com.github.saxypandabear.songrequests.util.JsonUtil.objectMapper
import com.typesafe.scalalogging.StrictLogging
import org.eclipse.jetty.websocket.api.{Session, WebSocketAdapter}

class TestResponseSocket extends WebSocketAdapter with StrictLogging {

  private val timer = new Timer("test-server")

  private val PONG_MESSAGE = """
                                 |{
                                 |  "type": "MESSAGE",
                                 |  "data": {
                                 |    "type": "PONG"
                                 |  }
                                 |}
                                 |""".stripMargin

  override def onWebSocketConnect(sess: Session): Unit = {
    super.onWebSocketConnect(sess)
    logger.info("Received connect event")
    timer.schedule(
        new RespondTimedTask(sess),
        0,
        WebSocketTestingUtil.sendFrequencyMs
    )
    WebSocketTestingUtil.onConnect.release()
  }

  override def onWebSocketClose(statusCode: Int, reason: String): Unit = {
    super.onWebSocketClose(statusCode, reason)
    logger.info(s"Received close event: Status=$statusCode | Reason=$reason")
    timer.cancel()
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
      case "PING"     => handlePingMessage(jsonMessage)
      case "LISTEN"   => handleListenMessage(jsonMessage)
      case "UNLISTEN" => handleUnlistenMessage(jsonMessage)
    }
  }

  private def handlePingMessage(jsonNode: JsonNode): Unit = {
    logger.info("Ping message received")
    // we send a PONG back
    getRemote.sendString(PONG_MESSAGE)
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

class RespondTimedTask(session: Session) extends TimerTask {

  /**
   * This needs to check the testing utility for whether we are allowed
   * to send messages, and if so, send so long as we have permits on the
   * semaphores. This should be able to mix redemption and reconnect
   * messages, in order to act like a real server (pseudo-chaos testing).
   * This should check for redeem events first, then check for
   */
  override def run(): Unit = {
    if (
        WebSocketTestingUtil.shouldSendRedeemEvent
          .get() && WebSocketTestingUtil.redeemEvents.availablePermits() > 0
    ) {
      WebSocketTestingUtil.redeemEvents.acquire()
      session.getRemote.sendStringByFuture(
          WebSocketTestingUtil.createRedeemEvent()
      )
    }
    if (
        WebSocketTestingUtil.shouldSendReconnectEvent
          .get() && WebSocketTestingUtil.reconnectEvents.availablePermits() > 0
    ) {
      WebSocketTestingUtil.reconnectEvents.acquire()
      session.getRemote.sendStringByFuture(
          WebSocketTestingUtil.createReconnectEvent()
      )
    }
  }
}
