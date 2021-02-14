package com.github.saxypandabear.songrequests

import org.junit.runner.RunWith
import org.scalatest.junit.JUnitRunner
import org.scalatest.{FlatSpec, Matchers}

// TODO: replace these with a UnitSpec extension, imported from a shared
//       scala library
@RunWith(classOf[JUnitRunner])
class FooTest extends FlatSpec with Matchers {
  it should "work" in {
    val foo = new Foo()
    foo.foo() should be("Bar")
  }
}
