package com.github.saxypandabear.songrequests.websocket

import com.github.saxypandabear.songrequests.ddb.ConnectionDataStore
import com.github.saxypandabear.songrequests.oauth.OauthTokenManagerFactory
import com.github.saxypandabear.songrequests.queue.SongQueue
import com.github.saxypandabear.songrequests.websocket.listener.WebSocketListener

/**
 * Factory for creating TwitchSocket objects
 */
class TwitchSocketFactory(
    clientId: String,
    clientSecret: String,
    refreshUri: String,
    tokenManagerFactory: OauthTokenManagerFactory,
    connectionDataStore: ConnectionDataStore,
    songQueue: SongQueue,
    listeners: Seq[WebSocketListener] = Seq.empty,
    pingFrequencyMs: Long = 60000
) {

  def create(channelId: String): TwitchSocket = {
    val tokenManager = tokenManagerFactory.create(
        clientId,
        clientSecret,
        channelId,
        refreshUri,
        connectionDataStore
    )

    new TwitchSocket(
        channelId,
        tokenManager,
        songQueue,
        listeners,
        pingFrequencyMs
    )
  }
}
