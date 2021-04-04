package com.github.saxypandabear.songrequests.intercept
import com.github.saxypandabear.songrequests.intercept.model.Message

import scala.collection.mutable

/**
 * Local implementation of an event interceptor, used for local development
 * and tests. Note that this exposes the internal queue, so that assertions
 * can be easily made on it without adding bloat to the overarching trait.
 * @param messageQueue initial message queue. defaults to an empty queue
 */
class InMemoryEventInterceptor(
    val messageQueue: mutable.Queue[Message] = mutable.Queue.empty
) extends EventInterceptor {

  /**
   * In an attempt to simulate the SQS configuration, this will poll
   * a single message.
   * @return a singleton list of the first element in the internal queue.
   *         if the internal queue is empty, return an empty list
   */
  override def poll(): Seq[Message] =
    messageQueue.synchronized {
      if (messageQueue.isEmpty) {
        Seq.empty
      } else {
        Seq(messageQueue.dequeue())
      }
    }

  // no-op
  override def shutdown(): Unit = {}

  /**
   * Clear all of the messages in the internal queue for testing
   */
  def flush(): Unit =
    messageQueue.synchronized {
      messageQueue.clear()
    }

  /**
   * For an in-memory implementation for testing, we need a mechanism in order
   * to actually synthesize messages to be consumed.
   * @param message Message to queue up for testing
   */
  def enqueue(message: Message): Unit =
    messageQueue += message
}
