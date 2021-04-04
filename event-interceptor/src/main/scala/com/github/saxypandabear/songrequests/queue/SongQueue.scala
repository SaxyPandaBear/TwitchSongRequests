package com.github.saxypandabear.songrequests.queue

import com.github.saxypandabear.songrequests.types.Types.ChannelId

/**
 * Something that accepts a message and channel ID, then "queues" the message input, which is
 * expected to be a song (Spotify URI)
 */
trait SongQueue {

  /**
   * Queue a song
   * @param channelId The Twitch channel ID that is spawned the request
   * @param spotifyUri The input Spotify URI to enqueue
   */
  def queue(channelId: ChannelId, spotifyUri: String): Unit

  /**
   * Stop receiving requests, cleaning up any necessary resources
   */
  def stop(): Unit

  /**
   * The queue URL that this class interacts with, i.e.: the SQS URL
   * @return The HTTP URL for the queue
   */
  def getQueueUrl: String
}
