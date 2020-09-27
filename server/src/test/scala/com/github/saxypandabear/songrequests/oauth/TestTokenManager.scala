package com.github.saxypandabear.songrequests.oauth

import scala.collection.mutable

class TestTokenManager extends OauthTokenManager {
    val refreshEvents = new mutable.HashMap[String, String]()
    val acceptedCredentials = Map("abc123" -> "foo")

    var accessToken: String = _

    /**
     * Retrieve an access token
     * @return an OAuth access token
     */
    override def getAccessToken: String = ???

    /**
     * Initiate a request to refresh the OAuth token, presumably because
     * the existing token is expired.
     * @param clientId     client ID for the application
     * @param clientSecret client secret for the application
     * @param refreshToken existing refresh token
     * @param uri          URI to request the token from
     * @return a POJO that represents the successful response from the authentication server
     */
    override def refresh(clientId: String, clientSecret: String, refreshToken: String, uri: String): OauthResponse = ???
}
