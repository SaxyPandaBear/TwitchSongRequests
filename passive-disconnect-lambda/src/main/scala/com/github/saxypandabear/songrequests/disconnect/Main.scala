package com.github.saxypandabear.songrequests.disconnect

import com.amazonaws.services.lambda.runtime.events.SQSEvent
import com.amazonaws.services.sqs.model.SendMessageRequest
import com.github.saxypandabear.songrequests.ddb.{
  DynamoDbConnectionDataStore,
  InMemoryConnectionDataStore
}
import com.github.saxypandabear.songrequests.metric.CloudWatchMetricCollector
import com.github.saxypandabear.songrequests.types.Types.ChannelId
import com.github.saxypandabear.songrequests.util.{
  AwsUtil,
  JsonUtil,
  ProjectProperties
}
import com.typesafe.scalalogging.LazyLogging

import java.util.concurrent.Executors
import scala.collection.JavaConverters._

/**
 * Lambda that receives messages from an SQS queue. The message contains a
 * Twitch channel ID. The lambda queries the Twitch API in order to see if
 * that channel is actively streaming, and if not, then we can set the status
 * of the channel to inactive in our database.
 */
class Main extends LazyLogging {
  private val DEACTIVATE_CONNECTION_METRIC = "Twitch-Channel-Not-Active"
  private val TWITCH_API_URI_KEY           = "twitch.api.uri"
  private val projectProperties            = new ProjectProperties().withSystemProperties()
  private val sqsClient                    = AwsUtil.createSqsClient(projectProperties)
  private val executor                     = Executors.newFixedThreadPool(3)
  private val metricCollector              = new CloudWatchMetricCollector(
      AwsUtil.createCloudWatchClient(projectProperties),
      executor
  )
  private val connectionDataStore          = projectProperties.getString("env") match {
    case Some("local") => new InMemoryConnectionDataStore()
    case Some(_)       =>
      new DynamoDbConnectionDataStore(
          AwsUtil.createDynamoDbClient(projectProperties)
      )
    case None          => new InMemoryConnectionDataStore()
  }

  /**
   * Main function for this lambda. Consumes the SQS event, checks Twitch API
   * if the channel is still live. If not, update DynamoDB to reflect that.
   * @param event consumed SQS event
   * @return The result of the execution, as a string message.
   */
  def checkConnections(event: SQSEvent): String =
    try {
      // TODO: try to parallelize this. I don't anticipate that this will
      //       actually help because it is most likely that the SQS event
      //       will only contain one message at a time.
      event.getRecords.asScala.par.foreach { message =>
        val channelId = extractChannelId(message.getBody)
        logger.info(
            "Checking if Twitch channel {} is still streaming",
            channelId
        )
        if (!isChannelStillActivelyStreaming(channelId)) {
          setChannelToInactive(channelId)
          metricCollector.emitCountMetric(
              DEACTIVATE_CONNECTION_METRIC,
              1.0,
              Map("channelId" -> channelId)
          )
        } else {
          // if the channel is still actively streaming, we have to consume
          // this message again. this SQS hook is most likely going to
          // automatically purge this message from the queue so we have to
          // re-drive it.
          putMessageBackIntoQueue(message)
        }
      }
      "Success"
    } catch {
      case e: Exception =>
        logger.error("Unexpected exception when processing SQS events", e)
        if (e.getMessage != null) {
          e.getMessage
        } else {
          s"Unhandled exception occurred: ${e.getClass.getSimpleName}"
        }
    }

  /**
   * Deserialize the JSON body of the message, and extract the channelId value
   * from it, returning that value. This assumes that the message contains
   * a "channelId" key
   * @param messageBody message body from SQS message
   * @return the value of the "channelId" key in the message
   */
  private def extractChannelId(messageBody: String): ChannelId =
    JsonUtil.objectMapper.readTree(messageBody).get("channelId").asText()

  /**
   * Makes an API call to the Twitch v5 API:
   * https://dev.twitch.tv/docs/v5/reference/streams#get-live-streams
   * If it responds with an active channel with the supplied channel ID, then
   * this indicates that the channel is currently live, and we should not
   * do work on it.
   * @param channelId Twitch channel ID
   * @return true if the Twitch channel is actively streaming, according to the
   *         Twitch API, false otherwise.
   */
  private def isChannelStillActivelyStreaming(channelId: ChannelId): Boolean =
    false // TODO: implement

  /**
   * Update the status in our database for this channel to be inactive.
   * @param channelId Twitch channel ID
   */
  private def setChannelToInactive(channelId: ChannelId): Unit = {
    connectionDataStore.updateConnectionStatus(channelId, "inactive")
    logger.info("Updated {} to inactive status", channelId)
  }

  // todo: implement
  /**
   * Re-drive an SQS message, at a later point in time, in order to check
   * again if the channel is currently active.
   * @param message SQS message that was consumed
   */
  private def putMessageBackIntoQueue(message: SQSEvent.SQSMessage): Unit = {
    val request = new SendMessageRequest().withQueueUrl("")
  }
}
