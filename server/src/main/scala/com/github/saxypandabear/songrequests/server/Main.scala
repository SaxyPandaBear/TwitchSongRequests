package com.github.saxypandabear.songrequests.server

import com.github.saxypandabear.songrequests.ddb.{
  ConnectionDataStore,
  DynamoDbConnectionDataStore,
  InMemoryConnectionDataStore
}
import com.github.saxypandabear.songrequests.metric.CloudWatchMetricCollector
import com.github.saxypandabear.songrequests.oauth.factory.TwitchOauthTokenManagerFactory
import com.github.saxypandabear.songrequests.queue.{
  InMemorySongQueue,
  SQSSongQueue,
  SongQueue
}
import com.github.saxypandabear.songrequests.util.AwsUtil._
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

import java.net.URI
import java.nio.file.Paths
import java.util.concurrent.Executors

/**
 * Main server entry point that stands up the Jetty server,
 * and all of the other infrastructure needed to run.
 */
object Main extends StrictLogging {
  // public scope for integration tests
  var orchestrator: ConnectionOrchestrator                = _
  private var server: Server                              = _
  private var metricsCollector: CloudWatchMetricCollector = _
  private var songQueue: SongQueue                        = _
  private var connectionDataStore: ConnectionDataStore    = _
  private var twitchSocketFactory: TwitchSocketFactory    = _
  private var region: String                              = _

  def main(args: Array[String] = Array.empty): Unit = {
    logger.info("Reading system and default application properties")
    val properties = new ProjectProperties()
      .withSystemProperties()
      .withResource("application.properties")
    for (filePath <- args) {
      logger.info("Loading override configuration from: {}", filePath)
      properties.withResourceAtPath(Paths.get(filePath))
    }

    logger.info(properties.toString())

    region = properties.getString("region").getOrElse("us-east-1")

    initMetricCollector(properties)
    initConnectionDataStore(properties)
    initSongQueue(properties)
    initOrchestrator(properties)
    initTwitchSocketFactory(properties)
    start(properties.getInteger("port").getOrElse(8080))
  }

  def start(port: Int): Unit = {
    logger.info("Server starting on port {}", port)
    val applicationBinder = new ApplicationBinder()
      .withImplementation(orchestrator, classOf[ConnectionOrchestrator])
      .withImplementation(twitchSocketFactory, classOf[TwitchSocketFactory])
    server = JettyUtil.build(port, applicationBinder)
    server.start()
  }

  def stop(): Unit = {
    logger.info("Server shutting down")
    server.stop()
    orchestrator.stop()
    songQueue.stop()
    connectionDataStore.stop()
    metricsCollector.stop()
  }

  private def initOrchestrator(projectProperties: ProjectProperties): Unit = {
    // if the properties defines a `twitch.url`, then we use that. if the
    // properties defines a `twitch.port`, then we know that this is used in a
    // local test and should be listening to localhost.
    // should fail fast if we don't have either.
    val twitchUri = if (projectProperties.has("twitch.url")) {
      new URI(projectProperties.get("twitch.url"))
    } else if (projectProperties.has("twitch.port")) {
      new URI(s"ws://localhost:${projectProperties.get("twitch.port")}")
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
    logger.info("Initializing metrics collector")
    val cloudWatch = createCloudWatchClient(projectProperties)
    metricsCollector = new CloudWatchMetricCollector(
        cloudWatch,
        Executors.newFixedThreadPool(
            projectProperties.getInteger("num.threads").getOrElse(10)
        )
    )
  }

  private def initSongQueue(projectProperties: ProjectProperties): Unit = {
    logger.info("Initializing song queue")
    songQueue = projectProperties.getString("env") match {
      case Some("local") => new InMemorySongQueue()
      case Some(_)       =>
        new SQSSongQueue(createSqsClient(projectProperties), metricsCollector)
      case None          => new InMemorySongQueue()
    }
  }

  private def initConnectionDataStore(
      projectProperties: ProjectProperties
  ): Unit = {
    logger.info("Initializing connection data store")
    connectionDataStore = projectProperties.getString("env") match {
      case Some("local") => new InMemoryConnectionDataStore()
      case Some(_)       =>
        new DynamoDbConnectionDataStore(createDynamoDbClient(projectProperties))
      case None          => new InMemoryConnectionDataStore()
    }
  }

  private def initTwitchSocketFactory(
      projectProperties: ProjectProperties
  ): Unit = {
    val clientId         = projectProperties.get("client.id")
    val clientSecret     = projectProperties.get("client.secret")
    val twitchRefreshUri = projectProperties.get("twitch.refresh.uri")
    val logListener      = new LoggingWebSocketListener()

    twitchSocketFactory = new TwitchSocketFactory(
        clientId,
        clientSecret,
        twitchRefreshUri,
        TwitchOauthTokenManagerFactory,
        connectionDataStore,
        songQueue,
        metricsCollector,
        Seq(logListener)
    )
  }
}
