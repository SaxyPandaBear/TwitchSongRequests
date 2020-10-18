package com.github.saxypandabear.songrequests.oauth

import java.util.UUID

import com.github.saxypandabear.songrequests.ddb.ConnectionDataStore

import scala.collection.mutable

class TestTokenManager(
    clientId: String,
    clientSecret: String,
    refreshToken: String,
    uri: String
) extends OauthTokenManager(clientId, clientSecret, refreshToken, uri) {
  var accessToken: String = _

  /**
   * Retrieve an access token
   * @return an OAuth access token
   */
  override def getAccessToken: String =
    TestTokenManager.clientIdsToTokens
      .getOrElseUpdate(clientId, UUID.randomUUID().toString)

  /**
   * Initiate a request to refresh the OAuth token, presumably because
   * the existing token is expired.
   * @return a POJO that represents the successful response from the authentication server
   */
  override def refresh(): OauthResponse =
    if (
        TestTokenManager.acceptedCredentials.exists(tuple =>
          tuple._1 == clientId && tuple._2 == clientSecret && tuple._3 == refreshToken
        )
    ) {
      // first update the token in our internal map so we can assert it properly
      // in tests
      val updatedToken = UUID.randomUUID().toString
      TestTokenManager.clientIdsToTokens.put(clientId, updatedToken)

      OauthResponse(updatedToken, refreshToken, "some-scope")
    } else {
      throw new RuntimeException
    }
}

object TestTokenManager {
  // this tracks the acceptable client ID -> secret -> refresh token grouping.
  // This should not change between tests
  val acceptedCredentials = Seq(("abc123", "foo", "bar"))

  // keep track of refresh events that happen in this manager
  val refreshEvents = new mutable.HashMap[String, String]()

  // maps a valid client ID (based on acceptedCredentials) to a generated token.
  val clientIdsToTokens = new mutable.HashMap[String, String]()

  def flush(): Unit = {
    refreshEvents.clear()
    clientIdsToTokens.clear()
  }
}

object TestTokenManagerFactory extends OauthTokenManagerFactory {

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
    val connection = connectionDataStore.getConnectionDetailsById(channelId)
    new TestTokenManager(
        clientId,
        clientSecret,
        connection.retrieveRefreshToken(),
        refreshUri
    )
  }
}
