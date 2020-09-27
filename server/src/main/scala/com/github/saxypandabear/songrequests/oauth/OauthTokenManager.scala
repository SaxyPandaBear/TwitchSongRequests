package com.github.saxypandabear.songrequests.oauth

trait OauthTokenManager {

    /**
     * Retrieve an access token
     * @return an OAuth access token
     */
    def getAccessToken: String

    /**
     * Initiate a request to refresh the OAuth token, presumably because
     * the existing token is expired.
     * @param clientId     client ID for the application
     * @param clientSecret client secret for the application
     * @param refreshToken existing refresh token
     * @param uri          URI to request the token from
     * @return a POJO that represents the successful response from the authentication server
     */
    def refresh(clientId: String,
                clientSecret: String,
                refreshToken: String,
                uri: String): OauthResponse
}
