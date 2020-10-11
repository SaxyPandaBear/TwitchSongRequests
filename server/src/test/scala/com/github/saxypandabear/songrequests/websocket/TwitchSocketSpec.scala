package com.github.saxypandabear.songrequests.websocket

import com.github.saxypandabear.songrequests.lib.UnitSpec

class TwitchSocketSpec extends UnitSpec {
  "Checking a valid user input Spotify URI" should "return true" in {
    val socket = new TwitchSocket("", null)
    socket.inputMatchesSpotifyUri("spotify:track:abc123") should be(true)
  }

  "Checking a valid user input Spotify URI that has extra whitespace" should "return true" in {
    val socket = new TwitchSocket("", null)
    socket.inputMatchesSpotifyUri("  spotify:track:abc123    \n") should be(
        true
    )
  }

  "Checking a user input that contains a valid Spotify URI but has other characters" should "return false" in {
    val socket = new TwitchSocket("", null)
    socket.inputMatchesSpotifyUri("before spotify:track:abc123") should be(
        false
    )
    socket.inputMatchesSpotifyUri("spotify:track:abc123 \nafter") should be(
        false
    )
  }

  "Checking an invalid input, like an album URI" should "return false" in {
    val socket = new TwitchSocket("", null)
    socket.inputMatchesSpotifyUri("spotify:album:foobarbaz987") should be(false)
  }

  "Checking a null Spotify URI" should "return false" in {
    val socket = new TwitchSocket("", null)
    socket.inputMatchesSpotifyUri(null) should be(false)
  }

  "Checking an empty string" should "return false" in {
    val socket = new TwitchSocket("", null)
    socket.inputMatchesSpotifyUri("") should be(false)
  }
}
