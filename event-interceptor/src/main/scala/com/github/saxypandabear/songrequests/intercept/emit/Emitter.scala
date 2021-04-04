package com.github.saxypandabear.songrequests.intercept.emit

import com.github.saxypandabear.songrequests.intercept.model.Message

trait Emitter {

  /**
   * Perform an action in reaction to a message that is received from the
   * parent interceptor
   * @param message message to emit
   */
  def emit(message: Message): Unit
}
