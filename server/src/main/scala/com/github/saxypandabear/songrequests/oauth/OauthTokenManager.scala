package com.github.saxypandabear.songrequests.oauth

abstract class OauthTokenManager(
    clientId: String,
    clientSecret: String,
    refreshToken: String,
    uri: String
) {

  /**
   * Retrieve an access token
   * @return an OAuth access token
   */
  def getAccessToken: String

  /**
   * Initiate a request to refresh the OAuth token, presumably because
   * the existing token is expired.
   * @return a POJO that represents the successful response from the authentication server
   */
  def refresh(): OauthResponse
}
