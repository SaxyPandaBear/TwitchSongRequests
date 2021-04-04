package com.github.saxypandabear.songrequests.intercept

import com.github.saxypandabear.songrequests.intercept.model.Message

trait EventInterceptor {

  /**
   * Poll for events that are available to act on
   * @return list of available messages
   */
  def poll(): Seq[Message]

  /**
   * Stop collection and clean up auxiliary resources
   */
  def shutdown(): Unit
}
