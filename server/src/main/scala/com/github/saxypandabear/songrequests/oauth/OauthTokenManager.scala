package com.github.saxypandabear.songrequests.oauth

/**
 * A class that handles the OAuth token for Twitch. Note that this does not
 * need to be generic enough to also handle Spotify authentication, because
 * none of that is handled here.
 * @param clientId     application client id
 * @param clientSecret application client secret
 * @param refreshToken refresh token for this specific instance. this is why
 *                     we cannot share a single OAuth token manager for the
 *                     entire application. Each client that is listening to
 *                     Twitch has been granted a different authorization from a
 *                     different user.
 * @param uri          URI that this should use to re-authenticate.
 */
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
