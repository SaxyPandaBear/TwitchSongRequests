package com.github.saxypandabear.songrequests.websocket

import com.github.saxypandabear.songrequests.ddb.ConnectionDataStore
import com.github.saxypandabear.songrequests.metric.CloudWatchMetricCollector
import com.github.saxypandabear.songrequests.oauth.factory.OauthTokenManagerFactory
import com.github.saxypandabear.songrequests.queue.SongQueue
import com.github.saxypandabear.songrequests.types.Types.ChannelId
import com.github.saxypandabear.songrequests.websocket.listener.WebSocketListener

/**
 * Factory for creating TwitchSocket objects
 * @param clientId Application Twitch client ID
 * @param clientSecret Application Twitch client secret
 * @param refreshUri Refresh URI for OAuth tokens
 * @param tokenManagerFactory OAuth manager factory for creating/refreshing
 *                            customer Twitch OAuth tokens
 * @param connectionDataStore Interface with database
 * @param songQueue Interface with queue
 * @param metricCollector CloudWatch metric client
 * @param listeners Sequence of socket listeners to include for each socket
 *                  created. Defaults to an empty Seq
 * @param pingFrequencyMs configurable time in milliseconds for the frequency
 *                        of server pings for each socket. Defaults to 1 minute
 */
class TwitchSocketFactory(
    clientId: String,
    clientSecret: String,
    refreshUri: String,
    tokenManagerFactory: OauthTokenManagerFactory,
    connectionDataStore: ConnectionDataStore,
    songQueue: SongQueue,
    metricCollector: CloudWatchMetricCollector,
    listeners: Seq[WebSocketListener] = Seq.empty,
    pingFrequencyMs: Long = 60000
) {

  /**
   * Create a TwitchSocket for a given channel ID
   * @param channelId Twitch channel ID
   * @return a new TwitchSocket implementation that listens to the given
   *         Twitch channel
   */
  def create(channelId: ChannelId): TwitchSocket = {
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
        metricCollector,
        listeners,
        pingFrequencyMs
    )
  }
}
