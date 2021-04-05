package com.github.saxypandabear.songrequests.spotify

import com.amazonaws.services.dynamodbv2.model.ResourceNotFoundException
import com.amazonaws.services.lambda.runtime.events.SQSEvent
import com.fasterxml.jackson.databind.ObjectMapper
import com.github.saxypandabear.songrequests.ddb.model.Connection
import com.github.saxypandabear.songrequests.ddb.{
  ConnectionDataStore,
  DynamoDbConnectionDataStore,
  InMemoryConnectionDataStore
}
import com.github.saxypandabear.songrequests.spotify.model.{
  Device,
  SpotifyDevicesResponse
}
import com.github.saxypandabear.songrequests.util.{
  AwsUtil,
  JsonUtil,
  ProjectProperties
}
import com.typesafe.scalalogging.LazyLogging
import org.apache.http.client.HttpClient
import org.apache.http.client.methods.{HttpGet, HttpPost}
import org.apache.http.client.utils.URIBuilder
import org.apache.http.impl.client.HttpClientBuilder

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
  val KEY_API_URL       = "SpotifyApiUrl"
  val KEY_OAUTH_URL     = "SpotifyOauthUrl"

  val projectProperties: ProjectProperties     =
    new ProjectProperties().withSystemProperties()
  val objectMapper: ObjectMapper               = JsonUtil.objectMapper
  val httpClient: HttpClient                   = HttpClientBuilder.create().build()
  val connectionDataStore: ConnectionDataStore = initDataStore(
      projectProperties
  )

  val clientId: String          = projectProperties.get(KEY_CLIENT_ID)
  val clientSecret: String      = projectProperties.get(KEY_CLIENT_SECRET)
  val oauthUrl: String          = projectProperties.get(KEY_OAUTH_URL)
  val spotifyApiBaseUrl: String = projectProperties.get(KEY_API_URL)

  /**
   * Main Lambda entry-point that accepts the SQS event message(s) and does
   * work on them.
   * @param event SQS event received by Lambda
   * @return empty string on success
   */
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
        val accessToken  = spotifyToken.get("access_token").asText()
        val refreshToken = spotifyToken.get("refresh_token").asText()

        val deviceOpt = findActiveComputer(accessToken)
        deviceOpt.map(d => queueSong(accessToken, d, spotifyUri))
      }
    }
    ""
  }

  /**
   * Call the Spotify API for this user, listing all of the available devices.
   * Filter the response down to just the active, computer devices, and return
   * the first device that matches the criteria.
   * Ref: https://developer.spotify.com/documentation/web-api/reference/#endpoint-get-a-users-available-devices
   * @param accessToken User's Spotify OAuth 2.0 access token
   * @return optionally, the first found active computer device for this user.
   */
  def findActiveComputer(accessToken: String): Option[Device] = {
    val uriBuilder = new URIBuilder(s"$spotifyApiBaseUrl/devices")
    val request    = new HttpGet(uriBuilder.build())
    request.setHeader("Accept", "application/json")
    request.setHeader("Authorization", s"Bearer $accessToken")

    val httpResponse = httpClient.execute(request)
    if (httpResponse.getStatusLine.getStatusCode != 200) {
      logger.error("Something blew up") // TODO: make this more robust
      None
    }
    val responseObj  = objectMapper.readValue(
        httpResponse.getEntity.getContent,
        classOf[SpotifyDevicesResponse]
    )
    responseObj.devices.find(d => d.isActive && d.deviceType == "Computer")
  }

  /**
   * Given a device, queue a Spotify URI to that device
   * for the given user.
   * Ref: https://developer.spotify.com/documentation/web-api/reference/#endpoint-add-to-queue
   * @param accessToken User OAuth access token
   * @param device      Found active computer device to queue song on
   * @param spotifyUri  Input Spotify URI
   */
  def queueSong(
      accessToken: String,
      device: Device,
      spotifyUri: String
  ): Unit = {
    val request = new HttpPost(
        s"$spotifyApiBaseUrl/queue?uri=$spotifyUri&device_id=${device.id}"
    )
    request.setHeader("Accept", "application/json")
    request.setHeader("Authorization", s"Bearer $accessToken")
    httpClient.execute(request)
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
    val requestBody = Map[String, String](
        "grant_type"    -> "refresh_token",
        "refresh_token" -> refreshToken,
        "client_id"     -> clientId,
        "client_secret" -> clientSecret
    )
    val uriBuilder  = new URIBuilder(oauthUrl)
    for ((k, v) <- requestBody)
      uriBuilder.addParameter(k, v)
    val request  = new HttpPost(uriBuilder.build())
    request.setHeader("Accept", "application/json")
    request.setHeader("Content-Type", "application/x-www-form-urlencoded")
    // TODO: this will become obsolete (hopefully) by migrating to secrets
    // manager to handle OAuth tokens, so leaving this unfinished for now.
    val response = httpClient.execute(request)
  }

  private def initDataStore(
      projectProperties: ProjectProperties
  ): ConnectionDataStore =
    projectProperties.getString("env") match {
      case Some("local") => new InMemoryConnectionDataStore()
      case Some(_)       =>
        new DynamoDbConnectionDataStore(
            AwsUtil.createDynamoDbClient(projectProperties)
        )
      case None          => new InMemoryConnectionDataStore()
    }
}
