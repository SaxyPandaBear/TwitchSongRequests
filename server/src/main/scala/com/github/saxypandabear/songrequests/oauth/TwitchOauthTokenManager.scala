package com.github.saxypandabear.songrequests.oauth

import com.github.saxypandabear.songrequests.ddb.ConnectionDataStore
import com.github.saxypandabear.songrequests.util.{HttpUtil, JsonUtil}

/**
 * Class that manages an OAuth token for Twitch. Twitch OAuth requests are
 * handled slightly differently than Spotify, which is why this is separated.
 * @param clientId     Application client ID used to authenticate against Twitch
 * @param clientSecret Application client secret used to authenticate against Twitch
 * @param channelId    Twitch channel ID that is associated with the token that this class manages
 * @param refreshUri   URI target for requesting a refresh token
 * @param dataStore    Interface to perform operations on the database
 */
class TwitchOauthTokenManager(
    clientId: String,
    clientSecret: String,
    channelId: String,
    refreshUri: String,
    refreshToken: String,
    dataStore: ConnectionDataStore
) extends OauthTokenManager(clientId, clientSecret, refreshToken, refreshUri) {
  private val connection = dataStore.getConnectionDetailsById(channelId)

  /**
   * Retrieve an access token
   * @return an OAuth access token
   */
  def getAccessToken: String = connection.twitchAccessToken()

  /**
   * Performs the token refresh, and also persists the change to DynamoDB
   * @return
   */
  override def refresh(): OauthResponse = {
    val response =
      requestNewToken(clientId, clientSecret, refreshToken, refreshUri)
    dataStore.updateTwitchOAuthToken(channelId, response.accessToken)
    response
  }

  private def requestNewToken(
      clientId: String,
      clientSecret: String,
      refreshToken: String,
      uri: String
  ): OauthResponse =
    HttpUtil.withAutoClosingHttpClient { httpClient =>
      val request  = httpClient
        .POST(uri)
        .param("grant_type", "refresh_token")
        .param("client_id", clientId)
        .param("client_secret", clientSecret)
        .param("refresh_token", refreshToken)
      val response = request.send()
      if (response.getStatus < 300) {
        JsonUtil.objectMapper
          .readValue(response.getContentAsString, classOf[OauthResponse])
      } else {
        // TODO: make this error handling better
        throw new RuntimeException(
            s"Refresh request responded with status ${response.getStatus}"
        )
      }
    }
}

object HttpOauthTokenManagerFactory extends OauthTokenManagerFactory {

  /**
   * Create some implementation of an OAuth token manager.
   * @param clientId            application client id
   * @param clientSecret        application client secret
   * @param channelId           Twitch channel ID
   * @param refreshUri          URI for re-authentication
   * @param connectionDataStore database wrapper that stores the bulk of
   *                            connection information
   * @return an implementation of an OAuth token manager
   */
  override def create(
      clientId: String,
      clientSecret: String,
      channelId: String,
      refreshUri: String,
      connectionDataStore: ConnectionDataStore
  ): OauthTokenManager = {
    // just need to extract the refresh token from the database
    // TODO: because of how this is written, there are 2 database reads.
    //       look into refactoring this so we only need to read once on
    //       initialization instead of twice to optimize.
    val refreshToken = connectionDataStore
      .getConnectionDetailsById(channelId)
      .twitchRefreshToken()

    new TwitchOauthTokenManager(
        clientId,
        clientSecret,
        channelId,
        refreshUri,
        refreshToken,
        connectionDataStore
    )
  }
}
