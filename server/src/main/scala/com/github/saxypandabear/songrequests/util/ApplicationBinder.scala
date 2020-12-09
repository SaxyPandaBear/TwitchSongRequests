package com.github.saxypandabear.songrequests.util

import org.glassfish.jersey.internal.inject.AbstractBinder

import scala.collection.mutable

/**
 * Binder class that is used for dependency injection
 */
class ApplicationBinder extends AbstractBinder {
  val implementationsToContracts =
    new mutable.ArrayBuffer[(AnyRef, Class[_ <: Any])]()

  override def configure(): Unit =
    for ((impl, clazz) <- implementationsToContracts)
      bind(impl).to(clazz)

  def withImplementation(
      impl: AnyRef,
      clazz: Class[_ <: Any]
  ): ApplicationBinder = {
    implementationsToContracts += ((impl, clazz))
    this
  }
}
