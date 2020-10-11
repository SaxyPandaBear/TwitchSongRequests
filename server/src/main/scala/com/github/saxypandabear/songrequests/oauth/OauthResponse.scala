package com.github.saxypandabear.songrequests.oauth

import com.fasterxml.jackson.annotation.JsonProperty

case class OauthResponse(
    @JsonProperty("access_token") accessToken: String,
    @JsonProperty("refresh_token") refreshToken: String,
    @JsonProperty("scope") scope: String
) {
  def this() {
    this("", "", "")
  }
}
