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
import org.eclipse.jetty.server.Server
import org.scalatest.BeforeAndAfterEach
import org.scalatest.concurrent.Eventually
import org.scalatest.mockito.MockitoSugar
import org.scalatest.time.{Millis, Span}

// wow that's a long name
class RoundRobinConnectionOrchestratorIntegrationSpec
    extends UnitSpec
    with RotatingTestPort
    with BeforeAndAfterEach
    with Eventually {

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
  }

  override def afterEach(): Unit =
    server.stop()

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
    orchestrator.connectionsToClients.values should contain theSameElementsAs Seq(
        Seq("a", "c", "e"),
        Seq("b", "d")
    )
  }

  private def initOrchestrator(numSockets: Int): Unit =
    orchestrator = new RoundRobinConnectionOrchestrator(
        metricCollector,
        new URI(s"ws://localhost:$port"),
        numSockets
    )
}
