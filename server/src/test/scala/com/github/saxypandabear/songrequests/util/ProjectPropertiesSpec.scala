package com.github.saxypandabear.songrequests.util

import com.github.saxypandabear.songrequests.UnitSpec

class ProjectPropertiesSpec extends UnitSpec {
    "Calling toMap on an empty properties object" should "return an empty map" in {
        val properties = new ProjectProperties()
        properties.toMap should be(empty)
    }

    "Setting a value in the map when the properties object is empty" should
        "put the value in the map" in {
        val properties = new ProjectProperties()
        properties should have size 0

        properties.setValue("foo", "bar")
        properties should have size 1
        properties.get("foo") should be("bar")
    }

    "Setting a value in the map that clashes with an existing key" should
        "replace the old value in the map" in {
        val properties = new ProjectProperties()
        properties.setValue("foo", "bar")
        properties.get("foo") should be("bar")
        properties.setValue("foo", "foo")
        properties.get("foo") should be("foo")
    }
}
