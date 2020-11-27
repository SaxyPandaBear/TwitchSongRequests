package com.github.saxypandabear.songrequests.server

import com.github.saxypandabear.songrequests.util.ProjectProperties
import com.github.saxypandabear.songrequests.websocket.orchestrator.{
  ConnectionOrchestrator,
  RoundRobinConnectionOrchestrator
}
import com.typesafe.scalalogging.StrictLogging
import org.eclipse.jetty.server.Server

/**
 * Main server entry point that stands up the Jetty server,
 * and all of the other infrastructure needed to run.
 */
object Main extends StrictLogging {
  private var server: Server                       = _
  private var orchestrator: ConnectionOrchestrator = _

  def main(args: Array[String]): Unit = {
    val properties = new ProjectProperties()
      .withSystemProperties()
      .withResource("application.properties")
    logger.info("Starting server")
    for ((k, v) <- properties)
      logger.info("{} = {}", k, v)

    start(properties.getInteger("port").getOrElse(8080))
  }

  def start(port: Int): Unit = {
    orchestrator = new RoundRobinConnectionOrchestrator()
    server = JettyUtil.build(port)
    server.start()
  }

  def stop(): Unit = {
    server.stop()
    orchestrator.stop()
  }
}
