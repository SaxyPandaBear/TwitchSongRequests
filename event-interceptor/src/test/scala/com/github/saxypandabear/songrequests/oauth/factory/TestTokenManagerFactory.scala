package com.github.saxypandabear.songrequests.oauth.factory

import com.github.saxypandabear.songrequests.ddb.model.Connection
import com.github.saxypandabear.songrequests.ddb.{
  ConnectionDataStore,
  InMemoryConnectionDataStore
}
import com.github.saxypandabear.songrequests.oauth.{
  OauthTokenManager,
  TestTokenManager
}
import com.github.saxypandabear.songrequests.util.JsonUtil.objectMapper
import com.typesafe.scalalogging.LazyLogging

object TestTokenManagerFactory
    extends OauthTokenManagerFactory
    with LazyLogging {

  // use this to leverage Scala's case class functionality in order to simply
  // copy the base object and change just the channel ID
  private val baseConnectionObj = objectMapper.readValue(
      getClass.getClassLoader.getResourceAsStream(
          "test-json/connection-active.json"
      ),
      classOf[Connection]
  )

  /**
   * Create some implementation of an OAuth token manager.
   * Note that for this test factory, if the channel ID does not exist in the
   * data store, we expect to generate one. Doing it here abstracts away the
   * work in the tests, and in practice, we expect the channel ID to exist
   * when it reaches this service.
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
    if (!connectionDataStore.isInstanceOf[InMemoryConnectionDataStore]) {
      throw new IllegalArgumentException(
          "Test token manager expects a local connection data store implementation"
      )
    }

    // if we don't have this channel ID yet, then create it and store it
    // before continuing
    if (!connectionDataStore.hasConnectionDetails(channelId)) {
      connectionDataStore
        .asInstanceOf[InMemoryConnectionDataStore]
        .putConnectionDetails(
            channelId,
            baseConnectionObj.copy(channelId = channelId)
        )
    }
    val connection = connectionDataStore.getConnectionDetailsById(channelId)
    new TestTokenManager(
        clientId,
        clientSecret,
        connection.twitchRefreshToken(),
        refreshUri
    )
  }
}
