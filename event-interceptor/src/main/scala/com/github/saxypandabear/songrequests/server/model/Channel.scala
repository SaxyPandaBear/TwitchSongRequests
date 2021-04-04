package com.github.saxypandabear.songrequests.server.model

import com.github.saxypandabear.songrequests.types.Types.ChannelId

case class Channel(channelId: ChannelId) {
  def this() {
    this("")
  }
}
