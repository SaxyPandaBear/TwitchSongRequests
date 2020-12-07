package com.github.saxypandabear.songrequests.server

import java.net.URI
import java.nio.file.Paths
import java.util.concurrent.Executors

import com.amazonaws.client.builder.AwsClientBuilder.EndpointConfiguration
import com.amazonaws.client.builder.AwsSyncClientBuilder
import com.amazonaws.services.cloudwatch.{
  AmazonCloudWatch,
  AmazonCloudWatchClientBuilder
}
import com.amazonaws.services.dynamodbv2.{
  AmazonDynamoDB,
  AmazonDynamoDBClientBuilder
}
import com.amazonaws.services.sqs.{AmazonSQS, AmazonSQSClientBuilder}
import com.github.saxypandabear.songrequests.ddb.{
  ConnectionDataStore,
  DynamoDbConnectionDataStore,
  InMemoryConnectionDataStore
}
import com.github.saxypandabear.songrequests.metric.CloudWatchMetricCollector
import com.github.saxypandabear.songrequests.oauth.factory.OauthTokenManagerFactory
import com.github.saxypandabear.songrequests.queue.{
  InMemorySongQueue,
  SQSSongQueue,
  SongQueue
}
import com.github.saxypandabear.songrequests.util.{
  ApplicationBinder,
  ProjectProperties
}
import com.github.saxypandabear.songrequests.websocket.TwitchSocketFactory
import com.github.saxypandabear.songrequests.websocket.listener.LoggingWebSocketListener
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
  private var server: Server                                     = _
  private[server] var orchestrator: ConnectionOrchestrator       = _
  private var metricsCollector: CloudWatchMetricCollector        = _
  private var songQueue: SongQueue                               = _
  private var connectionDataStore: ConnectionDataStore           = _
  private var oauthTokenManagerFactory: OauthTokenManagerFactory = _
  private var region: String                                     = _

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

    region = properties.getString("region").getOrElse("us-east-1")

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
    val cloudWatchBuilder =
      AmazonCloudWatchClientBuilder.standard().withRegion(region)
    setLocalStackUrlIfPresent(
        cloudWatchBuilder,
        "cloudwatch.url",
        projectProperties
    ).build()
  }

  private def createSqsClient(
      projectProperties: ProjectProperties
  ): AmazonSQS = {
    val sqsBuilder = AmazonSQSClientBuilder.standard().withRegion(region)
    setLocalStackUrlIfPresent(sqsBuilder, "sqs.url", projectProperties).build()
  }

  private def createDynamoDbClient(
      projectProperties: ProjectProperties
  ): AmazonDynamoDB = {
    val dynamoDbBuilder =
      AmazonDynamoDBClientBuilder.standard().withRegion(region)
    setLocalStackUrlIfPresent(
        dynamoDbBuilder,
        "dynamodb.url",
        projectProperties
    ).build()
  }

  private def initOrchestrator(projectProperties: ProjectProperties): Unit = {
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

  private def initMetricCollector(
      projectProperties: ProjectProperties
  ): Unit = {
    val cloudWatch = createCloudWatchClient(projectProperties)
    metricsCollector = new CloudWatchMetricCollector(
        cloudWatch,
        Executors.newFixedThreadPool(
            projectProperties.getInteger("num.threads").getOrElse(10)
        )
    )
  }

  private def initSongQueue(projectProperties: ProjectProperties): Unit =
    songQueue = projectProperties.getString("env") match {
      case Some("test") => new InMemorySongQueue()
      case Some(_)      =>
        new SQSSongQueue(createSqsClient(projectProperties), metricsCollector)
      case None         => new InMemorySongQueue()
    }

  private def initConnectionDataStore(
      projectProperties: ProjectProperties
  ): Unit =
    connectionDataStore = projectProperties.getString("env") match {
      case Some("test") => new InMemoryConnectionDataStore()
      case Some(_)      =>
        new DynamoDbConnectionDataStore(createDynamoDbClient(projectProperties))
      case None         => new InMemoryConnectionDataStore()
    }

  private def initTwitchOauthTokenManagerFactory(
      projectProperties: ProjectProperties
  ): Unit = {}

  private def createTwitchSocketFactory(
      projectProperties: ProjectProperties
  ): TwitchSocketFactory = {
    val clientId         = projectProperties.get("client.id")
    val clientSecret     = projectProperties.get("client.secret")
    val twitchRefreshUri = projectProperties.get("twitch.refresh.uri")
    val logListener      = new LoggingWebSocketListener()
    null // TODO: fix me
  }

  private def setLocalStackUrlIfPresent[
      Builder <: T,
      Type,
      T <: AwsSyncClientBuilder[Builder, Type]
  ](builder: T, key: String, projectProperties: ProjectProperties): T = {
    projectProperties.getString(key).foreach { url =>
      val endpointConfiguration = new EndpointConfiguration(url, region)
      builder.setEndpointConfiguration(endpointConfiguration)
    }

    builder
  }
}
