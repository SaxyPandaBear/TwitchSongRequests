package com.github.saxypandabear.songrequests.util

import com.github.saxypandabear.songrequests.lib.UnitSpec

import java.util.Properties

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

  "Adding from a Properties object" should "add new key value pairs, and update existing ones" in {
    val properties = new ProjectProperties()
    properties.setValue("foo", "bar")

    val props = new Properties()
    props.setProperty("foo", "baz")
    props.setProperty("bar", "foo")

    // this modifies our existing object, as well as returns it
    properties.`with`(props)

    properties should have size 2
    properties.get("foo") should be("baz")
    properties.get("bar") should be("foo")
  }

  "Adding from a properties file" should "add new key value pairs, and update existing ones" in {
    val properties = new ProjectProperties()
    properties.setValue("foo", "bar")

    val pathToPropertiesFile = "prop-test.properties"
    // this modifies our existing object, as well as returns it
    properties.withResource(pathToPropertiesFile)

    properties should have size 2
    properties.get("foo") should be("baz")
    properties.get("bar") should be("foo")
  }

  "Getting a boolean value" should "cast properly" in {
    val properties = new ProjectProperties()
    properties.setValue("foo", "true")
    properties.setValue("bar", "false")
    properties.setValue("baz", "not a boolean")

    val foo = properties.getBoolean("foo")
    foo should be(defined)
    foo.get should be(true)

    val bar = properties.getBoolean("bar")
    bar should be(defined)
    bar.get should be(false)

    properties.getBoolean("pickle") should not be defined
  }

  it should "throw an exception if the value cannot be parsed as as a boolean value" in {
    val properties = new ProjectProperties()
    properties.setValue("foo", "something")
    a[IllegalArgumentException] should be thrownBy properties.getBoolean("foo")
  }

  it should "return an empty Optional if the key does not exist in the properties map" in {
    val properties = new ProjectProperties()
    properties.getBoolean("foo") should not be defined
  }

  "Checking for a key that exists" should "return true" in {
    val properties = new ProjectProperties()
    properties.setValue("foo", "bar")

    properties.has("foo") should be(true)
  }

  "Checking for a key that does not exist" should "return false" in {
    val properties = new ProjectProperties()

    properties.has("foo") should be(false)
  }

  "Checking for a null or empty key" should "return false" in {
    val properties = new ProjectProperties()
    properties.setValue("foo", "bar")

    properties.has("") should be(false)
    properties.has(null) should be(false)
  }

  "The string representation of ProjectProperties" should "scrub out sensitive data" in {
    val projectProperties = new ProjectProperties()
    val offendingKey1     = "some_key"
    val offendingKey2     = "Some_mixed_CASE_SeCrEt"
    projectProperties.setValue(offendingKey1, "super-secret-value")
    projectProperties.setValue("foo", "bar")
    projectProperties.setValue(offendingKey2, "another super secret value")

    projectProperties.toString().startsWith("Project Properties:\n") should be(
        true
    )
    // don't need to compare against the whole string, because the iteration
    // of the hash map is not guaranteed to maintain any consistent ordering
    projectProperties
      .toString()
      .contains(s"\t$offendingKey1 : ${ProjectProperties.MASKED_VALUE}\n")
    projectProperties
      .toString()
      .contains(s"\t$offendingKey2 : ${ProjectProperties.MASKED_VALUE}\n")
    projectProperties.toString().contains("\tfoo: bar\n")
  }
}
