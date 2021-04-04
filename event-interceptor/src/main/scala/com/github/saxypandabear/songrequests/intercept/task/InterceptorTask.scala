package com.github.saxypandabear.songrequests.intercept.task

import com.github.saxypandabear.songrequests.intercept.EventInterceptor
import com.github.saxypandabear.songrequests.intercept.emit.Emitter
import com.github.saxypandabear.songrequests.intercept.receive.Receiver
import com.typesafe.scalalogging.LazyLogging

import java.util.concurrent.atomic.AtomicBoolean

class InterceptorTask(
    eventInterceptor: EventInterceptor,
    emitters: Seq[Emitter] = Seq.empty,
    receivers: Seq[Receiver] = Seq.empty
) extends Thread
    with LazyLogging {

  private val IdleWaitTimeMs: Long = 1000 * 60
  private val isRunning            = new AtomicBoolean(true)

  override def run(): Unit =
    try while (true) { // keep the thread alive
      while (isRunning.get()) {}
      // if we aren't running, then busy-wait until we should act again.
      // TODO: figure out how to trigger restarting the thread task
      Thread.sleep(IdleWaitTimeMs)
    } catch { // TODO: Introduce way to kill off thread intentionally
      case e: Exception =>
        logger.error("Exception encountered while polling for messages", e)
    }
}
