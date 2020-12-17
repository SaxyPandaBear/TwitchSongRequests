package com.github.saxypandabear.songrequests.queue

import com.github.saxypandabear.songrequests.types.Types.ChannelId

/**
 * Something that accepts a message and channel ID, then "queues" the message input, which is
 * expected to be a song (Spotify URI)
 */
trait SongQueue {
  def queue(channelId: ChannelId, spotifyUri: String): Unit
  def stop(): Unit
  def getQueueUrl: String
}
