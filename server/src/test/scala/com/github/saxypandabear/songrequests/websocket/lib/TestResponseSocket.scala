package com.github.saxypandabear.songrequests.websocket.lib

import java.util.concurrent.atomic.AtomicBoolean
import java.util.{Timer, TimerTask}

import com.fasterxml.jackson.databind.JsonNode
import com.github.saxypandabear.songrequests.util.JsonUtil.objectMapper
import com.typesafe.scalalogging.StrictLogging
import org.eclipse.jetty.websocket.api.{Session, WebSocketAdapter}

class TestResponseSocket extends WebSocketAdapter with StrictLogging {

  private val timer = new Timer("test-server")

  private val PONG_MESSAGE = """
                                 |{
                                 |  "type": "PONG"
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
  // use these flags to determine whether to send a redeem event or a
  // reconnect. this is so there is at least some deterministic approach to
  // sending responses from the server. these are specifically to cover
  // the case where we should send both redeem and reconnect events. we
  // don't want to send both at the same time, and want some determinism.
  // This really is just a mechanism for an XOR.
  // This gives precedence to the redeem events, since that defaults to true.
  // In the case where we only do reconnect events, this task will not send
  // a message in the first invocation, and that's okay.
  private val doSendRedeem    = new AtomicBoolean(true)
  private val doSendReconnect = new AtomicBoolean(false)

  /**
   * This needs to check the testing utility for whether we are allowed
   * to send messages, and if so, send so long as we have permits on the
   * semaphores. This should be able to mix redemption and reconnect
   * messages, in order to act like a real server (pseudo-chaos testing).
   * This should check for redeem events first, then check for
   */
  override def run(): Unit = {
    var sentSomething = false
    if (
        WebSocketTestingUtil.shouldSendRedeemEvent
          .get() && WebSocketTestingUtil.redeemEvents.availablePermits() > 0
    ) {
      if (doSendRedeem.get()) {
        WebSocketTestingUtil.redeemEvents.acquire()
        session.getRemote
          .sendStringByFuture(
              WebSocketTestingUtil.createRedeemEvent()
          )
          .get()
        if (WebSocketTestingUtil.startSending.availablePermits() == 0) {
          WebSocketTestingUtil.startSending.release()
        }
        doSendRedeem.getAndSet(false) // next iteration, don't send again
        sentSomething = true
      } else {
        doSendRedeem.getAndSet(true) // if we didn't send this time, we do send next time
      }
    }
    if (
        WebSocketTestingUtil.shouldSendReconnectEvent
          .get() && WebSocketTestingUtil.reconnectEvents.availablePermits() > 0
    ) {
      if (doSendReconnect.get()) {
        WebSocketTestingUtil.reconnectEvents.acquire()
        session.getRemote
          .sendStringByFuture(
              WebSocketTestingUtil.createReconnectEvent()
          )
          .get()
        if (WebSocketTestingUtil.startSending.availablePermits() == 0) {
          WebSocketTestingUtil.startSending.release()
        }
        doSendReconnect.getAndSet(false) // next iteration, don't send again
        sentSomething = true
      } else {
        doSendReconnect.getAndSet(true) // if we didn't send this time, we do send next time
      }
    }

    // if we did not send anything, then we can state that we're done sending
    WebSocketTestingUtil.doneSending.getAndSet(true)
  }
}
