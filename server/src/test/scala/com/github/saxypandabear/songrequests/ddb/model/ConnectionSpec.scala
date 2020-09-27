package com.github.saxypandabear.songrequests.ddb.model

import com.github.saxypandabear.songrequests.lib.UnitSpec
import com.github.saxypandabear.songrequests.util.JsonUtil.objectMapper
import org.scalatest.BeforeAndAfterEach

class ConnectionSpec extends UnitSpec with BeforeAndAfterEach {
    private val resourceFilePath = "test-json/connection-active.json"
    private var connection: Connection = _

    override def beforeEach(): Unit = {
        connection = objectMapper.readValue(
            getClass.getClassLoader.getResourceAsStream(resourceFilePath),
            classOf[Connection]
        )
    }

    "Getting the Twitch access token" should "work" in {
        connection.retrieveAccessToken should be("abc123")
    }

    "Getting the Twitch refresh token" should "work" in {
        connection.retrieveRefreshToken should be("foo")
    }

    "Setting the Twitch access token to a new value" should
        "be reflected when getting the access token afterwards" in {
        val someAccessToken = "hello, world"
        connection.retrieveAccessToken should be("abc123")
        connection.updateAccessToken(someAccessToken)
        connection.retrieveAccessToken should be(someAccessToken)
    }
}
