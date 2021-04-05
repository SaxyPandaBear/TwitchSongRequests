package com.github.saxypandabear.songrequests.intercept.task

import com.github.saxypandabear.songrequests.intercept.EventInterceptor
import com.github.saxypandabear.songrequests.intercept.emit.Emitter
import com.github.saxypandabear.songrequests.intercept.receive.Receiver
import com.typesafe.scalalogging.LazyLogging

import java.util.concurrent.atomic.AtomicBoolean

/**
 * Threaded task that polls for messages, forwards that message to it's
 * emitters, and reacts to input from it's receivers. This threaded task polls
 * on a set frequency for messages. It is possible to toggle the task on/off,
 * and when the task thread is idle, it
 * @param eventInterceptor event interceptor to poll from
 * @param emitters list of emitters to forward the interceptor message through
 * @param receivers list of receivers to accept external messages from
 * @param taskFrequencyMs frequency to poll from interceptor
 * @param idleFrequencyMs frequency to wait while task is idle
 */
class InterceptorTask(
    eventInterceptor: EventInterceptor,
    emitters: Seq[Emitter] = Seq.empty,
    receivers: Seq[Receiver] = Seq.empty,
    taskFrequencyMs: Long = 10000,
    idleFrequencyMs: Long = 60000
) extends Thread
    with LazyLogging {
  private val isRunning = new AtomicBoolean(true)

  override def run(): Unit =
    try while (true) { // keep the thread alive
      while (isRunning.get())
        try {
          val messages = eventInterceptor.poll()
          for (msg <- messages)
            for (emitter <- emitters)
              emitter.emit(msg)
          Thread.sleep(taskFrequencyMs)
        } catch {
          case e: Exception =>
            logger.error("An exception occurred while polling for messages", e)
        }
      // if we aren't running, then busy-wait until we should act again.
      // TODO: figure out how to trigger restarting the thread task
      Thread.sleep(idleFrequencyMs)
    } catch { // TODO: Introduce way to kill off thread intentionally
      case e: Exception =>
        logger.error("Thread killed", e)
    }
}
