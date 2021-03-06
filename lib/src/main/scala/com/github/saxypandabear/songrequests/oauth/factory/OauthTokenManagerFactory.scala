package com.github.saxypandabear.songrequests.oauth.factory

import com.github.saxypandabear.songrequests.ddb.ConnectionDataStore
import com.github.saxypandabear.songrequests.oauth.OauthTokenManager
import com.github.saxypandabear.songrequests.types.Types.ChannelId

trait OauthTokenManagerFactory {

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
  def create(
      clientId: String,
      clientSecret: String,
      channelId: ChannelId,
      refreshUri: String,
      connectionDataStore: ConnectionDataStore
  ): OauthTokenManager
}
