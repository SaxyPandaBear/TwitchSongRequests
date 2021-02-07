package com.github.saxypandabear.songrequests.spotify

import com.amazonaws.services.lambda.runtime.events.SQSEvent
import com.fasterxml.jackson.databind.ObjectMapper
import com.github.saxypandabear.songrequests.util.{JsonUtil, ProjectProperties}

import scala.collection.JavaConverters._

/**
 * This class expects messages from SQS, which should include
 * the Spotify URI in the record body, and the channel ID as
 * a message attribute.
 */
class Handler {
  val KEY_CHANNEL_ID                       = "channelId"
  val projectProperties: ProjectProperties =
    new ProjectProperties().withSystemProperties()
  val objectMapper: ObjectMapper           = JsonUtil.objectMapper

  def handle(event: SQSEvent): String = {
    for (record <- event.getRecords.asScala) {
      val channelId  =
        record.getMessageAttributes.get(KEY_CHANNEL_ID).getStringValue
      val spotifyUri = record.getBody
    }
    ""
  }
}
