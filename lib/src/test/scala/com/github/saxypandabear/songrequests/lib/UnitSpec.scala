package lib

import org.junit.runner.RunWith
import org.scalatest.junit.JUnitRunner
import org.scalatest.{FlatSpec, Matchers}

/**
 * This base class is used in order to provide a simple workaround for
 * Gradle to know to run our unit tests. All unit test classes should
 * extend from this, or provide their own workaround method in order for
 * Gradle to pick them up in the test task.
 *
 * See: https://stackoverflow.com/questions/18823855/cant-run-scalatest-with-gradle
 */
@RunWith(classOf[JUnitRunner])
abstract class UnitSpec extends FlatSpec with Matchers
