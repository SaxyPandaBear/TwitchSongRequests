package com.github.saxypandabear.songrequests.websocket.integration

import java.net.URI
import java.util.concurrent.Executors

import com.github.saxypandabear.songrequests.ddb.InMemoryConnectionDataStore
import com.github.saxypandabear.songrequests.lib.{
  DummyAmazonCloudWatch,
  RotatingTestPort,
  UnitSpec
}
import com.github.saxypandabear.songrequests.metric.CloudWatchMetricCollector
import com.github.saxypandabear.songrequests.oauth.TestTokenManagerFactory
import com.github.saxypandabear.songrequests.queue.InMemorySongQueue
import com.github.saxypandabear.songrequests.websocket.TwitchSocketFactory
import com.github.saxypandabear.songrequests.websocket.lib.WebSocketTestingUtil
import com.github.saxypandabear.songrequests.websocket.listener.{
  LoggingWebSocketListener,
  TestingWebSocketListener
}
import com.github.saxypandabear.songrequests.websocket.orchestrator.RoundRobinConnectionOrchestrator
import com.typesafe.scalalogging.LazyLogging
import org.eclipse.jetty.server.Server
import org.scalatest.BeforeAndAfterEach
import org.scalatest.concurrent.Eventually
import org.scalatest.time.{Millis, Span}

// wow that's a long name
class RoundRobinConnectionOrchestratorIntegrationSpec
    extends UnitSpec
    with RotatingTestPort
    with BeforeAndAfterEach
    with Eventually
    with LazyLogging {

  private var orchestrator: RoundRobinConnectionOrchestrator = _
  private val songQueue                                      = new InMemorySongQueue()
  private val dataStore                                      = new InMemoryConnectionDataStore()
  private val logListener                                    = new LoggingWebSocketListener()
  private val testListener                                   = new TestingWebSocketListener()
  private var twitchSocketFactory: TwitchSocketFactory       = _
  private var metricCollector: CloudWatchMetricCollector     = _
  private var testCloudWatchClient: DummyAmazonCloudWatch    = _
  private var server: Server                                 = _
  private val executor                                       = Executors.newFixedThreadPool(1)

  override def beforeEach(): Unit = {
    super.beforeEach()

    testCloudWatchClient = new DummyAmazonCloudWatch
    metricCollector = new CloudWatchMetricCollector(
        testCloudWatchClient,
        executor
    )
    twitchSocketFactory = new TwitchSocketFactory(
        "foo",
        "bar",
        "baz",
        TestTokenManagerFactory,
        dataStore,
        songQueue,
        metricCollector,
        Seq(logListener, testListener)
    )

    server = WebSocketTestingUtil.build(port)
    server.start()

    logger.info("Starting test with server hosted on port {}", port)
  }

  override def afterEach(): Unit = {
    server.stop()
    WebSocketTestingUtil.reset()
  }

  "Stopping the orchestrator" should "stop all of the internal WebSocket clients" in {
    val uri              = new URI(s"ws://localhost:$port")
    // in case calling stop() on a WebSocket client regresses and causes an
    // exception to be thrown, create a new orchestrator to test this
    val someOrchestrator =
      new RoundRobinConnectionOrchestrator(metricCollector, uri)
    someOrchestrator.connectionsToClients.keys.forall(_.isRunning) should be(
        true
    )
    someOrchestrator.stop()
    eventually(timeout(Span(100, Millis))) {
      someOrchestrator.connectionsToClients.keys.forall(_.isStopped) should be(
          true
      )
    }
  }

  "Initializing an orchestrator" should "not accept a number of sockets that is less than 1" in {
    val numSockets = 0
    val exception  = intercept[IllegalArgumentException] {
      initOrchestrator(numSockets)
    }
    exception should have message s"Orchestrator misconfigured - requires at least 1 client, but received $numSockets"
  }

  "An orchestrator with no 'active' connections" should "return a map that is not empty, but contains empty values" in {
    val numSockets  = 5
    initOrchestrator(numSockets)
    val connections = orchestrator.connectionsToClients

    connections should have size numSockets
    connections.values.foreach(_.isEmpty should be(true))
  }

  "Connecting channels" should "connect to clients in a round-robin fashion" in {
    val numSockets = 3
    initOrchestrator(numSockets)

    // by having 3 channels to connect, we expect each channel ID to be
    // connected to a different client
    val channelIds = Seq("a", "b", "c")

    channelIds.foreach(orchestrator.connect(_, twitchSocketFactory.create))
    orchestrator.connectionsToClients.values.foreach(_ should have size 1)
    orchestrator.connectionsToClients.values.flatten should contain theSameElementsAs channelIds
  }

  "Connecting channels" should "deterministically choose which client to connect to" in {
    val numSockets = 2
    initOrchestrator(numSockets)

    // having 2 clients but 5 channels, we should expect the orchestrator to
    // deterministically choose which client each channel should connect to
    val channelIds = Seq("a", "b", "c", "d", "e")

    channelIds.foreach(orchestrator.connect(_, twitchSocketFactory.create))
    orchestrator.connectionsToClients.values.flatten should contain theSameElementsAs channelIds
    orchestrator
      .indexedWebSocketConnections(0)
      ._2
      .map(_.channelId) should contain theSameElementsAs Seq("a", "c", "e")
    orchestrator
      .indexedWebSocketConnections(1)
      ._2
      .map(_.channelId) should contain theSameElementsAs Seq("b", "d")
  }

  "Connecting to many channels" should "eventually cause the orchestrator to reach its allowed capacity" in {
    val numSockets     = 2
    val numConnections = 2
    initOrchestrator(numSockets, numConnections)

    // there are only 2 connections allowed for each socket, and 2 sockets,
    // so having 5 channels to connect should cause the orchestrator to reach
    // capacity.
    val channelIdsWithIndexes = Seq("a", "b", "c", "d", "e").zipWithIndex
    val indexAtCapacity       = channelIdsWithIndexes.size - 1

    for ((id, index) <- channelIdsWithIndexes) {
      // we should successfully connect for all except the last one
      orchestrator.connect(id, twitchSocketFactory.create) should be(
          index != indexAtCapacity
      )
      if (index == indexAtCapacity) {
        orchestrator.atCapacity should be(true)
      } else {
        orchestrator.atCapacity should be(false)
      }
    }
  }

  "Disconnecting a channel from the orchestrator" should "remove the connection" in {
    // going to use the PING metrics to validate that two of the clients
    // disconnected. the ping frequency will help with this
    val frequencyMs         = 25
    val twitchSocketFactory = new TwitchSocketFactory(
        clientId = "foo",
        clientSecret = "bar",
        refreshUri = "baz",
        tokenManagerFactory = TestTokenManagerFactory,
        connectionDataStore = dataStore,
        songQueue = songQueue,
        metricCollector = metricCollector,
        listeners = Seq(logListener, testListener),
        pingFrequencyMs = frequencyMs
    )

    initOrchestrator(2)

    // splitting these up just to make assertions later simpler without
    // performing any extra splicing
    val remove = Seq("a", "b")
    val remain = Seq("c", "d", "e")
    remove.foreach(orchestrator.connect(_, twitchSocketFactory.create))
    remain.foreach(orchestrator.connect(_, twitchSocketFactory.create))

    // disconnecting "a" and "b" should leave us with (c, e) and (d), because
    // the provisioning is deterministic (see above test on determinism)
    remove.par.foreach(orchestrator.disconnect(_))

    // Our test listener captures all of the PONG events from the server,
    // per client. We aren't removing the client sockets from the listener
    // when we disconnect, so there should still be 5 things in the map.
    // take a snapshot of the map right now, so that we can assert against
    // it later.
    val startCounts =
      Map(testListener.messageEvents.mapValues(_.length).toSeq: _*)
    startCounts should have size 5
    logger.info("Starting counts are: {}", startCounts.mkString(","))

    // after N pongs from the server, we should see clear divergence between
    // the counts for the remaining connections and the removed ones.
    val numPongs     = 10
    // the max we're going to allow for the number of pongs that we receive
    // for the removed connections
    val ceilingPongs = numPongs / 2
    eventually(timeout(Span(frequencyMs * (numPongs + 1), Millis))) {
      val (remainCounts, removeCounts) =
        testListener.messageEvents.mapValues(_.length).partition {
          case (channelId, _) => remain.contains(channelId)
        }

      logger.info("Starting: {}", startCounts.mkString(","))
      logger.info("Remaining: {}", remainCounts.mkString(","))
      logger.info("Removed: {}", removeCounts.mkString(","))

      for ((channelId, count) <- remainCounts)
        count - startCounts(channelId) should be(numPongs +- 1)
      for ((channelId, count) <- removeCounts)
        count - startCounts(channelId) should be <= ceilingPongs
    }

    // make sure that the internal state reflects the disconnected channels
    orchestrator.connectionsToClients.values.flatten should contain theSameElementsAs remain
    orchestrator
      .indexedWebSocketConnections(0)
      ._2
      .map(_.channelId) should contain theSameElementsAs Seq("c", "e")
    orchestrator
      .indexedWebSocketConnections(1)
      ._2
      .map(_.channelId) should contain theSameElementsAs Seq("d")
  }

  "Disconnecting a channel from an orchestrator that is at capacity" should "free up capacity on the orchestrator" in {
    val numSockets     = 2
    val numConnections = 2
    initOrchestrator(numSockets, numConnections)

    // when the orchestrator attempts to connect to "e", it will fail because
    // we are at capacity. (there is already a test for this)
    // what we want to do is then disconnect one of the
    val toDisconnect   = "a"
    val toConnectAfter = "e"
    val others         = Seq("b", "c", "d")

    orchestrator.connect(toDisconnect, twitchSocketFactory.create) should be(
        true
    )
    others.foreach(
        orchestrator.connect(_, twitchSocketFactory.create) should be(true)
    )
    orchestrator.connect(toConnectAfter, twitchSocketFactory.create) should be(
        false
    )
    orchestrator.atCapacity should be(true)

    // now, we disconnect one. this should free up a spot for "e"
    orchestrator.disconnect(toDisconnect)
    orchestrator.atCapacity should be(false)

    orchestrator.connectionsToClients.values.flatten should contain theSameElementsAs others
    orchestrator
      .indexedWebSocketConnections(0)
      ._2
      .map(_.channelId) should contain theSameElementsAs Seq("c")
    orchestrator
      .indexedWebSocketConnections(1)
      ._2
      .map(_.channelId) should contain theSameElementsAs Seq("b", "d")
    orchestrator.connect(toConnectAfter, twitchSocketFactory.create) should be(
        true
    )
  }

  private def initOrchestrator(
      numSockets: Int,
      numConnections: Int = 40
  ): Unit =
    orchestrator = new RoundRobinConnectionOrchestrator(
        metricCollector,
        new URI(s"ws://localhost:$port"),
        numSockets,
        numConnections
    )
}
