package com.github.saxypandabear.songrequests.server

import java.net.URI
import java.nio.file.Paths
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
import com.github.saxypandabear.songrequests.websocket.TwitchSocketFactory
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
  private var server: Server                               = _
  private[server] var orchestrator: ConnectionOrchestrator = _

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

    initOrchestrator(properties)
    start(properties.getInteger("port").getOrElse(8080))
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

  private def createCloudWatchClient(
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

  private def initOrchestrator(projectProperties: ProjectProperties): Unit = {
    val cloudWatch       = createCloudWatchClient(projectProperties)
    val metricsCollector = new CloudWatchMetricCollector(
        cloudWatch,
        Executors.newFixedThreadPool(
            projectProperties.getInteger("num.threads").getOrElse(10)
        )
    )

    // if the properties defines a `twitch.url`, then we use that. if the
    // properties defines a `twitch.port`, then we know that this is used in a
    // local test and should be listening to localhost.
    // should fail fast if we don't have either.
    val twitchUri = if (projectProperties.has("twitch.url")) {
      new URI(projectProperties.get("twitch.url"))
    } else if (projectProperties.has("twitch.port")) {
      new URI(s"http://localhost:${projectProperties.get("twitch.port")}")
    } else {
      throw new RuntimeException(
          "Cannot start server because no Twitch server configuration set."
      )
    }

    orchestrator =
      new RoundRobinConnectionOrchestrator(metricsCollector, twitchUri)
  }

  private def createTwitchSocketFactory(
      projectProperties: ProjectProperties
  ): TwitchSocketFactory = {
    val clientId         = projectProperties.get("client.id")
    val clientSecret     = projectProperties.get("client.secret")
    val twitchRefreshUri = projectProperties.get("twitch.refresh.uri")

  }
}
