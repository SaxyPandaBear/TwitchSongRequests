package com.github.saxypandabear.songrequests.util

import java.util.Properties

import scala.collection.JavaConverters._
import scala.collection.mutable

/**
 * Stores configuration and properties, and provides does the necessary
 * casting in order to fetch strongly typed primitive configuration values.
 */
class ProjectProperties {
  private val internalMap = mutable.HashMap[String, String]()

  /**
   * Convenience method that transforms a snapshot of the properties
   * object into an immutable Map[String, String]. This notably does not
   * transform the values into things that they could be typed as.
   * For example, if there exists some `foo -> 42` in the internal map,
   * we do not return a map that includes a mapping of "foo" to the integer
   * value 42. It would result in `foo -> "42"`.
   *
   * @return an immutable map of keys to their raw string representations
   */
  def toMap: Map[String, String] = internalMap.toMap

  /**
   * Get the size of the properties object, which is just the size of the
   * internal map.
   *
   * @return the size of the internal map
   */
  def size: Int = internalMap.size

  /**
   * Exposes the set of keys that are defined in the properties object.
   *
   * @return a Seq of the internal map's keys
   */
  def keys: Seq[String] = internalMap.keys.toSeq

  /**
   * Exposes the set of values that are defined in the properties object.
   *
   * @return a Seq of the internal map's values
   */
  def values: Seq[String] = internalMap.values.toSeq

  /**
   * Simple setter that upserts a raw string value into the map
   *
   * @param key key to associate the raw value with
   * @param value string value to put in the map
   */
  def setValue(key: String, value: String): Unit =
    internalMap.put(key, value)

  /**
   * A convenience builder for adding a Properties object into
   * the internal map of this object.
   *
   * Leveraging the default way that Scala maps handle clashing keys,
   * this method will replace old values with the values found in the
   * input if the keys match.
   *
   * @param properties java.util.Properties object to include properties here
   * @return this properties object
   */
  def `with`(properties: Properties): ProjectProperties = {
    internalMap ++= properties.asScala
    this
  }

  /**
   * A convenience builder method that allows us to specify a resource file path,
   * like our application properties file, reads that file, and put the values into
   * our map.
   *
   * @param resourcePath path from `src/[context]/resources/`, where the "context"
   *                     is main when running the app, and test when running unit
   *                     tests
   * @return this properties object
   */
  def withResource(resourcePath: String): ProjectProperties = {
    val properties = new Properties()
    properties.load(getClass.getClassLoader.getResourceAsStream(resourcePath))
    internalMap ++= properties.asScala
    this
  }

  /**
   * A convenience builder method to inject System properties, like application
   * properties passed in via the command line:
   * `java Foo -Dbar=Baz`
   * as well as environment variables
   *
   * @return this properties object
   */
  def withSystemProperties(): ProjectProperties = {
    internalMap ++= System.getenv().asScala
    this
  }

  /**
   * A "raw" get method. This method is provided in place of a potential
   * `getString()` method because that would be redundant. This does not
   * perform any checks to ensure that the key exists in the map.
   *
   * Note: This is the only `get` function that does not wrap the output in
   *       an Option.
   *
   * @param key key to use in order to fetch data from the map
   * @return the raw string value associated with the input key
   * @throws NoSuchElementException if the key is not in the map
   */
  def get(key: String): String =
    internalMap(key)

  /**
   * Get a boolean value from the map, given the input key. This should throw
   * an exception when the value this tries to parse is not able to be parsed
   * into a boolean value.
   *
   * @param key key to use in order to fetch data from the map
   * @return Some(value) if the key exists in the map, and the raw value
   *         can be interpreted as a boolean, else None
   * @throws IllegalArgumentException when the input is not able to be parsed as a boolean
   */
  def getBoolean(key: String): Option[Boolean] =
    internalMap.get(key).map(v => v.toBoolean)
}
