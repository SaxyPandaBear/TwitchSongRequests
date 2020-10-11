package com.github.saxypandabear.songrequests.websocket

import java.util.concurrent.Semaphore

import com.fasterxml.jackson.databind.JsonNode
import org.eclipse.jetty.server.Server
import org.eclipse.jetty.servlet.ServletContextHandler

import scala.collection.mutable

object WebSocketTestingUtil {

  // keeps track of the "message types" that the Twitch Socket can send to the
  // server
  val acceptedMessageTypes = Set("PING", "LISTEN", "UNLISTEN")

  // keeps track of the channel IDs that are allowed to interact with the
  // server.
  // this helps to manage paths for testing, i.e.: which channel IDs trigger
  // error events, etc.
  val acceptedChannelIds = Set("abc123")
  // Stores server state that tracks which events occur when handling messages
  val pingMessages       = new mutable.ArrayBuffer[JsonNode]()
  val listenMessages     = new mutable.ArrayBuffer[JsonNode]()
  val unlistenMessages   = new mutable.ArrayBuffer[JsonNode]()
  // Locking mechanisms to block on events
  var onConnect          = new Semaphore(1)
  var onClose            = new Semaphore(1)
  var onError            = new Semaphore(1)
  var onMessage          = new Semaphore(1)

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
  }
}
