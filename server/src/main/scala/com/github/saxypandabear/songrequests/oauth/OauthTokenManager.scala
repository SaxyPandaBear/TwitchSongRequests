package com.github.saxypandabear.songrequests.oauth

import com.github.saxypandabear.songrequests.ddb.ConnectionDataStore
import com.github.saxypandabear.songrequests.util.JsonUtil
import org.eclipse.jetty.client.HttpClient

/**
 * Class that manages an OAuth token.
 * @param httpClient   HTTP client implementation (from Jetty)
 * @param dataStore    Interface to perform operations on the database
 * @param refreshUri   URI target for requesting a refresh token
 * @param clientId     app client id
 * @param clientSecret app client secret
 * @param accessToken  initial access token value
 * @param refreshToken refresh token
 */
class OauthTokenManager(httpClient: HttpClient,
                        dataStore: ConnectionDataStore,
                        refreshUri: String,
                        clientId: String,
                        clientSecret: String,
                        accessToken: String,
                        refreshToken: String) {

    def getAccessToken: String = accessToken

    /**
     * Perform an OAuth refresh of the token, given the input URI for the refresh
     */
    def refresh(channelId: String): Unit = {
        val token = requestNewToken()
        // need to update the Connection object in the data store with our new access token
        throw new UnsupportedOperationException("Implement me")
    }

    private def requestNewToken(): OauthResponse = {
        val request = httpClient.POST(refreshUri)
        val response = request.send()
        if (response.getStatus < 300) {
            JsonUtil.objectMapper.readValue(response.getContentAsString, classOf[OauthResponse])
        } else {
            // TODO: make this error handling better
            throw new RuntimeException(s"Refresh request responded with status ${response.getStatus}")
        }

    }
}
