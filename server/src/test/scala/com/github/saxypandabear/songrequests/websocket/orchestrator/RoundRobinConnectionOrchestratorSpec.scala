package com.github.saxypandabear.songrequests.websocket.orchestrator

import com.github.saxypandabear.songrequests.ddb.InMemoryConnectionDataStore
import com.github.saxypandabear.songrequests.lib.UnitSpec
import com.github.saxypandabear.songrequests.oauth.TestTokenManagerFactory
import com.github.saxypandabear.songrequests.queue.InMemorySongQueue
import org.scalatest.BeforeAndAfterEach

/**
 * This is mostly for testing the internal methods that are used, and thread safety.
 * The outer integration test is going to be the main source of testing
 * general functionality.
 */
class RoundRobinConnectionOrchestratorSpec
    extends UnitSpec
    with BeforeAndAfterEach {
  private var orchestrator: RoundRobinConnectionOrchestrator = _
  private val dataStore                                      = new InMemoryConnectionDataStore()
  private val clientId                                       = "foo"
  private val clientSecret                                   = "bar"
  private val refreshUri                                     = "baz"
  private val songQueue                                      = new InMemorySongQueue()

  override def afterEach(): Unit = {
    orchestrator.stop()
    dataStore.clear()
    songQueue.clear()
  }

  it should "not accept a number of sockets that is less than 1" in {
    val numSockets = 0
    val exception  = intercept[IllegalArgumentException] {
      initOrchestrator(numSockets)
    }
    exception should have message s"Orchestrator misconfigured - requires at least 1 client, but received $numSockets"
  }

  "An orchestrator with no active connections" should "return a map that is not empty, but contains empty values" in {
    val numSockets  = 5
    initOrchestrator(numSockets)
    val connections = orchestrator.activeConnections

    connections should have size numSockets
    for (channelIds <- connections.values)
      channelIds.isEmpty should be(true)
  }

  private def initOrchestrator(
      numSockets: Int
  ): Unit =
    orchestrator = new RoundRobinConnectionOrchestrator(
        clientId,
        clientSecret,
        refreshUri,
        dataStore,
        songQueue,
        TestTokenManagerFactory,
        numSockets
    )
}
