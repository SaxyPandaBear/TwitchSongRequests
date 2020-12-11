package com.github.saxypandabear.songrequests.ddb.model

import lib.UnitSpec
import com.github.saxypandabear.songrequests.util.JsonUtil.objectMapper
import org.scalatest.BeforeAndAfterEach

class ConnectionSpec extends UnitSpec with BeforeAndAfterEach {
  private val resourceFilePath       = "test-json/connection-active.json"
  private var connection: Connection = _

  override def beforeEach(): Unit =
    connection = objectMapper.readValue(
        getClass.getClassLoader.getResourceAsStream(resourceFilePath),
        classOf[Connection]
    )

  // this is actually tested already because it is the prerequisite
  // step to the rest of the tests, but just as another point of
  // code coverage
  "Reading a JSON file that represents a Connection object" should "work" in {
    val connectionObj = objectMapper.readValue(
        getClass.getClassLoader.getResourceAsStream(resourceFilePath),
        classOf[Connection]
    )

    connectionObj.channelId should be("1234567890")
    connectionObj.expires should be(9876543210L)
    connectionObj.connectionStatus should be("active")

    // we can validate the whole JSON string, but since this is just
    // a slight sanity check, I'm only asserting that the
    // access and refresh keys are correct
    val sessionObject = objectMapper.readTree(connectionObj.sess)
    val twitchToken   = sessionObject.get("accessKeys").get("twitchToken")
    twitchToken.get("access_token").asText() should be("abc123")
    twitchToken.get("refresh_token").asText() should be("foo")
    val spotifyToken  = sessionObject.get("accessKeys").get("spotifyToken")
    spotifyToken.get("access_token").asText() should be("321cba")
    spotifyToken.get("refresh_token").asText() should be("bar")
  }

  "Getting the Twitch access token" should "work" in {
    connection.twitchAccessToken should be("abc123")
  }

  "Getting the Twitch refresh token" should "work" in {
    connection.twitchRefreshToken should be("foo")
  }

  "Setting the Twitch access token to a new value" should
    "be reflected when getting the access token afterwards" in {
      val someAccessToken = "hello, world"
      connection.twitchAccessToken should be("abc123")
      connection.updateTwitchAccessToken(someAccessToken)
      connection.twitchAccessToken should be(someAccessToken)
    }

  "Calling toItem" should "map values to the proper keys in the DynamoDB item" in {
    val valueMap = connection.toValueMap
    valueMap("channelId").getS should be("1234567890")
    valueMap("expires").getN.toLong should be(9876543210L)
    valueMap("connectionStatus").getS should be("active")

    // same sanity check as above
    val sessionObject = objectMapper.readTree(valueMap("sess").getS)
    val twitchToken   = sessionObject.get("accessKeys").get("twitchToken")
    twitchToken.get("access_token").asText() should be("abc123")
    twitchToken.get("refresh_token").asText() should be("foo")
    val spotifyToken  = sessionObject.get("accessKeys").get("spotifyToken")
    spotifyToken.get("access_token").asText() should be("321cba")
    spotifyToken.get("refresh_token").asText() should be("bar")
  }
}
