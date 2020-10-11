package com.github.saxypandabear.songrequests.queue

import scala.collection.mutable

class TestingSongQueue extends SongQueue {

    val queued = new mutable.HashMap[String, mutable.ArrayBuffer[String]]()

    override def queue(channelId: String, spotifyUri: String): Unit = {

    }

    def clear(): Unit = {
        queued.clear()
    }
}
