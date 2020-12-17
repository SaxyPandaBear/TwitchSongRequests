package com.github.saxypandabear.songrequests.oauth.factory
import com.github.saxypandabear.songrequests.ddb.ConnectionDataStore
import com.github.saxypandabear.songrequests.oauth.{
  OauthTokenManager,
  TwitchOauthTokenManager
}
import com.github.saxypandabear.songrequests.types.Types.ChannelId

object TwitchOauthTokenManagerFactory extends OauthTokenManagerFactory {

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
      channelId: ChannelId,
      refreshUri: String,
      connectionDataStore: ConnectionDataStore
  ): OauthTokenManager = {
    // TODO: this does two database reads. if this becomes a bottleneck,
    //       need to refactor this
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
