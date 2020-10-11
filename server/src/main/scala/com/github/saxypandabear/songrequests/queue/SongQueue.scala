package com.github.saxypandabear.songrequests.queue

/**
 * Something that accepts a message and channel ID, then "queues" the message input, which is
 * expected to be a song (Spotify URI)
 */
trait SongQueue {
    def queue(channelId: String, spotifyUri: String): Unit
}
