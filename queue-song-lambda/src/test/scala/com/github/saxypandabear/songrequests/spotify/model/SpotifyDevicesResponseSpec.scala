package com.github.saxypandabear.songrequests.spotify.model

import com.github.saxypandabear.songrequests.lib.UnitSpec
import com.github.saxypandabear.songrequests.util.JsonUtil
import junit.framework.Assert.{assertEquals, assertFalse, assertTrue}

/**
 * JSON tests for the Spotify API response
 * https://developer.spotify.com/documentation/web-api/reference/#endpoint-get-a-users-available-devices
 */
class SpotifyDevicesResponseSpec extends UnitSpec {

  // A sanity check for Jackson annotations to map the
  // response field names to the corresponding POJO fields
  it should "parse the named fields correctly" in {
    val response = JsonUtil.objectMapper.readValue(
        getClass.getClassLoader.getResourceAsStream("devices-response.json"),
        classOf[SpotifyDevicesResponse]
    )
    assertTrue(response.devices.nonEmpty)
    val device   = response.devices.head
    assertEquals("5fbb3ba6aa454b5534c4ba43a8c7e8e45a63ad0e", device.id)
    assertFalse(device.isActive)
    assertTrue(device.isPrivateSession)
    assertFalse(device.isRestricted)
    assertEquals("My fridge", device.name)
    assertEquals("Computer", device.deviceType)
    assertEquals(100, device.volumePercent)
  }
}
