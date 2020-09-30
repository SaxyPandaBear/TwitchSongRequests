package com.github.saxypandabear.songrequests.oauth

import com.github.saxypandabear.songrequests.ddb.ConnectionDataStore
import com.github.saxypandabear.songrequests.ddb.model.Connection
import com.github.saxypandabear.songrequests.util.{HttpUtil, JsonUtil}

/**
 * Class that manages an OAuth token.
 * @param clientId     Application client ID used to authenticate against Twitch
 * @param clientSecret Application client secret used to authenticate against Twitch
 * @param channelId    Twitch channel ID that is associated with the token that this class manages
 * @param refreshUri   URI target for requesting a refresh token
 * @param dataStore    Interface to perform operations on the database
 */
class OauthTokenManagerImpl(clientId: String,
                            clientSecret: String,
                            channelId: String,
                            refreshUri: String,
                            refreshToken: String,
                            dataStore: ConnectionDataStore)
    extends OauthTokenManager(clientId, clientSecret, refreshToken, refreshUri) {
    private val connection = dataStore.getConnectionDetailsById(channelId)

    /**
     * Retrieve an access token
     * @return an OAuth access token
     */
    def getAccessToken: String = connection.retrieveAccessToken()

    /**
     * Performs the token refresh, and also persists the change to DynamoDB
     * @return
     */
    override def refresh(): OauthResponse = {
        val response = requestNewToken(clientId, clientSecret, refreshToken, refreshUri)
        dataStore.updateConnectionDetailsById(channelId, connection)
        response
    }

    private def requestNewToken(clientId: String, clientSecret: String, refreshToken: String, uri: String): OauthResponse = {
        HttpUtil.withAutoClosingHttpClient(httpClient => {
            val request = httpClient.POST(uri)
                .param("grant_type", "refresh_token")
                .param("client_id", clientId)
                .param("client_secret", clientSecret)
                .param("refresh_token", refreshToken)
            val response = request.send()
            if (response.getStatus < 300) {
                JsonUtil.objectMapper.readValue(response.getContentAsString, classOf[OauthResponse])
            } else {
                // TODO: make this error handling better
                throw new RuntimeException(s"Refresh request responded with status ${response.getStatus}")
            }
        })
    }
}
