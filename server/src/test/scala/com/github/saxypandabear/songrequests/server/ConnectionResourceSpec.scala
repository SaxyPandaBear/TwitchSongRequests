package com.github.saxypandabear.songrequests.server

import com.github.saxypandabear.songrequests.lib.{RotatingTestPort, UnitSpec}
import com.github.saxypandabear.songrequests.server.model.Channel
import com.github.saxypandabear.songrequests.util.JsonUtil
import io.restassured.RestAssured
import io.restassured.http.ContentType
import org.eclipse.jetty.server.Server
import org.scalatest.{BeforeAndAfterAll, BeforeAndAfterEach}

class ConnectionResourceSpec
    extends UnitSpec
    with BeforeAndAfterEach
    with BeforeAndAfterAll
    with RotatingTestPort {

  private var server: Server = _

  override def beforeAll(): Unit = {
    RestAssured.enableLoggingOfRequestAndResponseIfValidationFails()
    RestAssured.useRelaxedHTTPSValidation()
  }

  override def beforeEach(): Unit = {
    super.beforeEach()
    server = JettyUtil.build(port)
    server.start()
  }

  override def afterEach(): Unit = {
    super.afterEach()
    server.stop()
  }

  it should "be healthy" in {
    val response = RestAssured
      .get(s"http://localhost:$port/api/ping")
      .`then`()
      .extract()
      .body()
      .asString()
    response should be("pong")
  }

  "Sending a request to connect to a Twitch channel" should "work" in {
    val id      = "12345"
    val channel = Channel(id)

    RestAssured
      .`given`()
      .contentType(ContentType.JSON)
      .accept(ContentType.TEXT)
      .body(JsonUtil.objectMapper.writeValueAsString(channel))
      .post(s"http://localhost:$port/api/connect")
      .`then`()
      .assertThat()
      .statusCode(201)
      .and()
      .extract()
      .body()
      .asString() should be(s"Initiated connection to channel $id")
  }

  "Submitting a request to disconnect from a Twitch channel" should "return a 204 on success" in {
    val id = "12345"

    RestAssured
      .`given`()
      .contentType(ContentType.JSON)
      .accept(ContentType.TEXT)
      .pathParam("channel", id)
      .put(s"http://localhost:$port/api/disconnect/{channel}")
      .`then`()
      .assertThat()
      .statusCode(204)
  }
}
