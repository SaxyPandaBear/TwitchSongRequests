package com.github.saxypandabear.songrequests.queue

import com.github.saxypandabear.songrequests.types.Types.ChannelId
import com.typesafe.scalalogging.StrictLogging

import scala.collection.mutable

class InMemorySongQueue extends SongQueue with StrictLogging {

  private val lockObject = new Object()
  val queued             = new mutable.HashMap[String, mutable.ArrayBuffer[String]]()

  override def queue(channelId: ChannelId, spotifyUri: String): Unit = {
    logger.info(
        s"Received queue event: Channel = $channelId - URI = $spotifyUri"
    )
    lockObject.synchronized {
      val songs =
        queued.getOrElseUpdate(channelId, new mutable.ArrayBuffer[String]())
      songs += spotifyUri
    }
  }

  def clear(): Unit = {
    logger.info("Clearing")
    queued.clear()
  }

  // no-op. nothing needs to be done for this
  override def stop(): Unit = {}

  // the in-memory variant has no real URL to point to
  override def getQueueUrl: String = ""
}
