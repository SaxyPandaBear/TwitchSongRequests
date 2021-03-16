package com.github.saxypandabear.songrequests.websocket

import com.github.saxypandabear.songrequests.lib.UnitSpec

/**
 * Tests that don't require any of the integrations, like the
 * song queue, OAuth, or metrics.
 */
class TwitchSocketSpec extends UnitSpec {

  private val socket = new TwitchSocket("", null, null, null)

  "Checking a valid user input Spotify URI" should "return true" in {
    socket.inputMatchesSpotifyUri("spotify:track:abc123") should be(true)
  }

  "Checking a valid user input Spotify URI that has extra whitespace" should "return true" in {
    socket.inputMatchesSpotifyUri("  spotify:track:abc123    \n") should be(
        true
    )
  }

  "Checking a user input that contains a valid Spotify URI but has other characters" should "return false" in {
    socket.inputMatchesSpotifyUri("before spotify:track:abc123") should be(
        false
    )
    socket.inputMatchesSpotifyUri("spotify:track:abc123 \nafter") should be(
        false
    )
  }

  "Checking an invalid input, like an album URI" should "return false" in {
    socket.inputMatchesSpotifyUri("spotify:album:foobarbaz987") should be(false)
  }

  "Checking a null Spotify URI" should "return false" in {
    socket.inputMatchesSpotifyUri(null) should be(false)
  }

  "Checking an empty string" should "return false" in {
    socket.inputMatchesSpotifyUri("") should be(false)
  }

  "Checking a reward with the title 'TwitchSongRequests'" should "return true" in {
    socket.isSongRequest("This is a TwitchSongRequests reward") should be(true)
    socket.isSongRequest("TwitchSongRequests") should be(true)
  }

  "Checking a reward title that does not contain the exact phrase" should "return false" in {
    socket.isSongRequest("This has a song, but is not a request") should be(
        false
    )
    socket.isSongRequest("twitchsongrequests") should be(false)
    socket.isSongRequest("This is a song request reward") should be(true)
    socket.isSongRequest("SONG REQUEST ALL CAPS") should be(true)
    socket.isSongRequest(
        "Song\n request with a line break is invalid"
    ) should be(false)
  }

  "Checking a reward title that is null" should "return false" in {
    socket.isSongRequest(null) should be(false)
  }

  "Checking a reward title that is empty" should "return false" in {
    socket.isSongRequest("") should be(false)
  }
}
