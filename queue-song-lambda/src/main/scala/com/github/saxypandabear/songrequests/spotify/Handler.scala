package com.github.saxypandabear.songrequests.spotify

import com.amazonaws.services.dynamodbv2.model.ResourceNotFoundException
import com.amazonaws.services.lambda.runtime.events.SQSEvent
import com.fasterxml.jackson.databind.ObjectMapper
import com.github.saxypandabear.songrequests.ddb.model.Connection
import com.github.saxypandabear.songrequests.ddb.{
  ConnectionDataStore,
  DynamoDbConnectionDataStore
}
import com.github.saxypandabear.songrequests.util.{
  AwsUtil,
  JsonUtil,
  ProjectProperties
}
import com.typesafe.scalalogging.LazyLogging
import org.apache.http.client.HttpClient
import org.apache.http.client.methods.HttpPost
import org.apache.http.client.utils.URIBuilder
import org.apache.http.entity.StringEntity
import org.apache.http.impl.client.HttpClientBuilder

import java.nio.charset.StandardCharsets
import scala.collection.JavaConverters._

/**
 * This class expects messages from SQS, which should include
 * the Spotify URI in the record body, and the channel ID as
 * a message attribute.
 */
class Handler extends LazyLogging {
  val KEY_CHANNEL_ID    = "channelId"
  val KEY_CLIENT_ID     = "SpotifyClientId"
  val KEY_CLIENT_SECRET = "SpotifyClientSecret"
  val KEY_BASE_URL      = "spotify.url"

  val projectProperties: ProjectProperties     =
    new ProjectProperties().withSystemProperties()
  val objectMapper: ObjectMapper               = JsonUtil.objectMapper
  val httpClient: HttpClient                   = HttpClientBuilder.create().build()
  val connectionDataStore: ConnectionDataStore = initDataStore(
      projectProperties
  )

  val clientId: String     = projectProperties.get(KEY_CLIENT_ID)
  val clientSecret: String = projectProperties.get(KEY_CLIENT_SECRET)
  val baseUrl: String      = projectProperties.get(KEY_BASE_URL)

  def handle(event: SQSEvent): String = {
    for (record <- event.getRecords.asScala) {
      val channelId  =
        record.getMessageAttributes.get(KEY_CHANNEL_ID).getStringValue
      val spotifyUri = record.getBody

      val connectionDetails: Connection = {
        try connectionDataStore.getConnectionDetailsById(channelId)
        catch {
          case e: ResourceNotFoundException =>
            logger.error("{} does not exist", channelId, e)
            throw e
          case exception: Exception         =>
            logger.error(
                "Unhandled exception occurred while attempting to fetch connection details for {}",
                channelId,
                exception
            )
            throw exception
        }
      }
      if (connectionDetails.connectionStatus != "active") {
        logger.info(
            "{} is not connected to the main service. Dropping record {} from SQS",
            channelId,
            record.getMessageId
        )
      } else {
        val sessionObj   = objectMapper.readTree(connectionDetails.sess)
        val spotifyToken = sessionObj.get("accessKeys").get("spotifyToken")
        val accessToken  = spotifyToken.get("access_token")
        val refreshToken = spotifyToken.get("refresh_token")

      }
    }
    ""
  }

  /**
   * Update the refresh token by requesting a new token from Spotify,
   * and write the updated token into the DynamoDB table.
   * https://developer.spotify.com/documentation/general/guides/authorization-guide/
   * @param channelId    The Twitch channel ID (hash key for the DDB table)
   * @param clientId     Spotify client ID
   * @param clientSecret Spotify client secret
   * @param refreshToken Expired refresh token for this user's access
   */
  private def refreshSpotifyToken(
      channelId: String,
      clientId: String,
      clientSecret: String,
      refreshToken: String
  ): Unit = {
    // TODO: figure this out
    val requestBody = Map[String, String](
        "grant_type"    -> "refresh_token",
        "refresh_token" -> refreshToken,
        "client_id"     -> clientId,
        "client_secret" -> clientSecret
    )
    val uriBuilder  = new URIBuilder(baseUrl)
    for ((k, v) <- requestBody)
      uriBuilder.addParameter(k, v)
    val request = new HttpPost(uriBuilder.build())
    request.setHeader("Accept", "application.json")
    request.setHeader("Content-Type", "application/x-www-form-urlencoded")
    httpClient.execute(request)
  }

  private def initDataStore(
      projectProperties: ProjectProperties
  ): ConnectionDataStore =
    new DynamoDbConnectionDataStore(
        AwsUtil.createDynamoDbClient(projectProperties)
    )
}
