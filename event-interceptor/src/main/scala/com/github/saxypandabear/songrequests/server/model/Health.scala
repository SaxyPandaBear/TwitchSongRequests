package com.github.saxypandabear.songrequests.server.model

case class Health(healthy: Boolean) {
  def this() {
    this(true)
  }
}
