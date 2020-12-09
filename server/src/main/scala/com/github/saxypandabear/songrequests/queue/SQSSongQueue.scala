package com.github.saxypandabear.songrequests.queue

import java.util.concurrent.atomic.AtomicBoolean

import com.amazonaws.services.sqs.AmazonSQS
import com.amazonaws.services.sqs.model.{
  MessageAttributeValue,
  SendMessageRequest
}
import com.github.saxypandabear.songrequests.metric.CloudWatchMetricCollector
import com.typesafe.scalalogging.LazyLogging

import scala.collection.JavaConverters._

class SQSSongQueue(sqs: AmazonSQS, metricsCollector: CloudWatchMetricCollector)
    extends SongQueue
    with LazyLogging {
  val QUEUE_NAME            = "song-queue"
  val METRIC_TOTAL_ACK      = "song-requests-acknowledged"
  val METRIC_SEND_SUCCEEDED = "song-requests-succeeded"
  val METRIC_SEND_FAILED    = "song-requests-failed"

  private val running          = new AtomicBoolean(true)
  private var queueUrl: String = _

  init()

  override def queue(channelId: String, spotifyUri: String): Unit = {
    logger.info(
        "Received request from channel {} to queue Spotify song {}",
        channelId,
        spotifyUri
    )
    val tags = Map("channelId" -> channelId)
    metricsCollector.emitCountMetric(METRIC_TOTAL_ACK, 1, tags)

    val channelIdAttribute =
      new MessageAttributeValue()
        .withDataType("String")
        .withStringValue(channelId)
    val sendMessageRequest = new SendMessageRequest()
      .withQueueUrl(queueUrl)
      .withMessageBody(spotifyUri)
      .withMessageAttributes(
          Map[String, MessageAttributeValue](
              "channelId" -> channelIdAttribute
          ).asJava
      )
    try {
      val response = sqs.sendMessage(sendMessageRequest)
      logger.info(
          "Queueing {} for {} responded with message ID {}",
          spotifyUri,
          channelId,
          response.getMessageId
      )
      metricsCollector.emitCountMetric(METRIC_SEND_SUCCEEDED, 1, tags)
    } catch {
      case e: Exception =>
        logger.warn(
            "Exception occurred when attempting to queue {} for channel {}",
            spotifyUri,
            channelId,
            e
        )
        metricsCollector.emitCountMetric(METRIC_SEND_FAILED, 1, tags)
    }
  }

  // don't shutdown the metrics collector since it's a shared instance.
  override def stop(): Unit =
    running.synchronized {
      if (running.get) {
        running.getAndSet(false)
        logger.info("Shutting down SQS client")
        sqs.shutdown()
      }
    }

  override def getQueueUrl: String = queueUrl

  // a "startup" method that checks a queue already exists, or creates one if
  // needed.
  private def init(): Unit = {
    logger.info("Initializing song queue")
    val queues =
      try sqs.listQueues(QUEUE_NAME).getQueueUrls.asScala
      catch {
        case e: Exception =>
          logger.warn("Error occurred when trying to list queues", e)
          throw e
      }
    if (queues.nonEmpty) {
      // already have an existing queue to refer to. pick the first and move on
      queueUrl = queues.head
      logger.info("Found existing queue URL: {}", queueUrl)
    } else {
      // need to create a queue
      queueUrl =
        try sqs.createQueue(QUEUE_NAME).getQueueUrl
        catch {
          case e: Exception =>
            logger.warn("Error occurred when trying to create a new queue", e)
            throw e
        }
      logger.info("Created new SQS queue: {}", queueUrl)
    }
  }
}
