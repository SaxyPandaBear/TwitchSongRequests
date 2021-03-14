package com.github.saxypandabear.songrequests.spotify.model

import com.fasterxml.jackson.annotation.{
  JsonCreator,
  JsonIgnoreProperties,
  JsonProperty
}

// https://developer.spotify.com/documentation/web-api/reference/#endpoint-get-a-users-available-devices
case class SpotifyDevicesResponse(devices: Seq[Device]) {
  @JsonCreator
  def this() {
    this(Seq.empty)
  }
}

@JsonIgnoreProperties(ignoreUnknown = true)
case class Device(
    id: String,
    @JsonProperty("is_active") isActive: Boolean,
    @JsonProperty("is_private_session") isPrivateSession: Boolean,
    @JsonProperty("is_restricted") isRestricted: Boolean,
    name: String,
    @JsonProperty("type") deviceType: String,
    volumePercent: Int
)
