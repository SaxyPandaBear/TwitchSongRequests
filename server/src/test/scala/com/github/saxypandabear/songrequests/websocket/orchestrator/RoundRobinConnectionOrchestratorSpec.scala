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

  override def beforeEach(): Unit =
    orchestrator = new RoundRobinConnectionOrchestrator(
        clientId,
        clientSecret,
        refreshUri,
        dataStore,
        songQueue,
        TestTokenManagerFactory
    )

  override def afterEach(): Unit = {
    orchestrator.stop()
    dataStore.clear()
  }
}
