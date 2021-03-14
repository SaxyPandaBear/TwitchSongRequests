package com.github.saxypandabear.songrequests.spotify.model

import com.fasterxml.jackson.annotation.{
  JsonCreator,
  JsonIgnoreProperties,
  JsonProperty
}
import com.fasterxml.jackson.databind.PropertyNamingStrategy.SnakeCaseStrategy
import com.fasterxml.jackson.databind.annotation.JsonNaming

// https://developer.spotify.com/documentation/web-api/reference/#endpoint-get-a-users-available-devices
case class SpotifyDevicesResponse(devices: Seq[Device]) {
  @JsonCreator
  def this() {
    this(Seq.empty)
  }
}

@JsonIgnoreProperties(ignoreUnknown = true)
@JsonNaming(classOf[SnakeCaseStrategy])
case class Device(
    id: String,
    isActive: Boolean,
    isPrivateSession: Boolean,
    isRestricted: Boolean,
    name: String,
    @JsonProperty("type") deviceType: String,
    volumePercent: Int
)
