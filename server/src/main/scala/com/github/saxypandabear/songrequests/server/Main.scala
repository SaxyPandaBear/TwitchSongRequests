package com.github.saxypandabear.songrequests.server

import java.net.URI
import java.nio.file.{Files, Paths}
import java.util.concurrent.Executors

import com.amazonaws.client.builder.AwsClientBuilder.EndpointConfiguration
import com.amazonaws.services.cloudwatch.{
  AmazonCloudWatch,
  AmazonCloudWatchClientBuilder
}
import com.github.saxypandabear.songrequests.metric.CloudWatchMetricCollector
import com.github.saxypandabear.songrequests.util.{
  ApplicationBinder,
  ProjectProperties
}
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

  def main(args: Array[String] = Array.empty): Unit = {
    logger.info("Reading system and default application properties")
    val properties = new ProjectProperties()
      .withSystemProperties()
      .withResource("application.properties")
    for (filePath <- args) {
      logger.info("Loading override configuration from: {}", filePath)
      properties.withResourceAtPath(Paths.get(filePath))
    }

    logger.info("Starting server with following properties:")
    for ((k, v) <- properties)
      logger.info("{} = {}", k, v)

    val port = properties.getInteger("port").getOrElse(8080)
    initOrchestrator(properties, port)
    start(port)
  }

  def start(port: Int): Unit = {
    logger.info("Server starting on port {}", port)
    val applicationBinder = new ApplicationBinder()
      .withImplementation(orchestrator, classOf[ConnectionOrchestrator])
    server = JettyUtil.build(port, applicationBinder)
    server.start()
  }

  def stop(): Unit = {
    logger.info("Server shutting down")
    server.stop()
    orchestrator.stop()
  }

  private def getCloudWatchClient(
      projectProperties: ProjectProperties
  ): AmazonCloudWatch = {
    val cloudWatchBuilder = AmazonCloudWatchClientBuilder.standard()
    val region            = projectProperties.getString("region").getOrElse("us-east-1")
    cloudWatchBuilder.setRegion(region)

    projectProperties.getString("cloudwatch.url").foreach { url =>
      val endpoint = new EndpointConfiguration(url, region)
      cloudWatchBuilder.setEndpointConfiguration(endpoint)
    }

    cloudWatchBuilder.build()
  }

  private def initOrchestrator(
      projectProperties: ProjectProperties,
      port: Int
  ): Unit = {
    val cloudWatch       = getCloudWatchClient(projectProperties)
    val metricsCollector = new CloudWatchMetricCollector(
        cloudWatch,
        Executors.newFixedThreadPool(
            projectProperties.getInteger("num.threads").getOrElse(10)
        )
    )

    val twitchUri = new URI(
        projectProperties
          .getString("twitch.url")
          .getOrElse(s"http://localhost:$port")
    )
    orchestrator =
      new RoundRobinConnectionOrchestrator(metricsCollector, twitchUri)
  }
}
