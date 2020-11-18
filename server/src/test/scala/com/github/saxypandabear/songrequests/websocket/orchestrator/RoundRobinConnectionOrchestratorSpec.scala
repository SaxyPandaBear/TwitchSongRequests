package com.github.saxypandabear.songrequests.websocket.orchestrator

import com.github.saxypandabear.songrequests.lib.UnitSpec

/**
 * Tests that don't require the other integrations,
 * like a WebSocket server running.
 */
class RoundRobinConnectionOrchestratorSpec extends UnitSpec {
  "Rotating the position" should "work" in {
    val orchestrator = new RoundRobinConnectionOrchestrator(null, null)
    val positions    = Seq(0, 1, 2, 3)
    for (p <- positions)
      orchestrator.rotate(p) should be(p + 1)
  }

  it should "wrap the rotation around once it reaches the end" in {
    val maxNumSockets = 3
    val orchestrator  =
      new RoundRobinConnectionOrchestrator(null, null, maxNumSockets)
    val positions     = Seq(0, 1, 2, 3)
    for (p <- positions) {
      val expected = if (p == maxNumSockets) 0 else p + 1
      orchestrator.rotate(p) should be(expected)
    }
  }
}
