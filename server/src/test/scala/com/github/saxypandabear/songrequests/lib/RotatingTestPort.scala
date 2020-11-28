package com.github.saxypandabear.songrequests.lib

import scala.util.Random

/**
 * A test spec that requires a randomly rotating test port in order to confirm
 * operations on unique instances of connections (i.e.: listening on different ports for
 * unit testing with a dummy server)
 */
trait RotatingTestPort {
  protected var port: Int = _

  def beforeEach(): Unit =
    port = randomPort()

  private def randomPort(): Int = Random.nextInt(1000) + 5000
}
