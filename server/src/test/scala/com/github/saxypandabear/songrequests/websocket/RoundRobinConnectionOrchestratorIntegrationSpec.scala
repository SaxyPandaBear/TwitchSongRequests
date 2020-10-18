package com.github.saxypandabear.songrequests.websocket

import com.github.saxypandabear.songrequests.lib.{RotatingTestPort, UnitSpec}
import org.scalatest.BeforeAndAfterEach
import org.scalatest.concurrent.Eventually

// wow that's a long name
class RoundRobinConnectionOrchestratorIntegrationSpec
    extends UnitSpec
    with RotatingTestPort
    with BeforeAndAfterEach
    with Eventually {

  override def beforeEach(): Unit = {}

}
