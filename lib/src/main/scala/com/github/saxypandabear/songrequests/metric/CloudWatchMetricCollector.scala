package com.github.saxypandabear.songrequests.metric

import com.amazonaws.services.cloudwatch.AmazonCloudWatch
import com.amazonaws.services.cloudwatch.model.{
  Dimension,
  MetricDatum,
  PutMetricDataRequest,
  StandardUnit
}
import com.typesafe.scalalogging.LazyLogging

import java.util.Date
import java.util.concurrent.atomic.AtomicBoolean
import java.util.concurrent.{ExecutorService, TimeUnit}
import scala.collection.JavaConverters._

/**
 * Simple implementation that takes data and publishes to CloudWatch.
 * This should perform the operations asynchronously, because we expect
 * to share this object throughout the app, and can't let this metric
 * emitter be the bottleneck for processing data.
 * @param client          internal AWS SDK CloudWatch client
 * @param executorService thread pool executor to submit EmitMetricTasks to
 */
class CloudWatchMetricCollector(
    client: AmazonCloudWatch,
    executorService: ExecutorService
) {
  private val running = new AtomicBoolean(true)

  /**
   * Check if the metric collector is active
   * @return true if the collector is running, false otherwise
   */
  def isRunning: Boolean = running.get()

  /**
   * Emit a metric to CloudWatch.
   * @param name Name of the CW metric
   * @param value Numeric value for the metric
   * @param tags Optional tags to include on the metric. Defaults to an empty Map
   */
  def emitCountMetric(
      name: String,
      value: Double,
      tags: Map[String, String] = Map.empty
  ): Unit =
    if (isRunning) {
      executorService.submit(new EmitMetricTask(client, name, value, tags))
    }

  /**
   * Stop emitting metrics and tear down all of the resources
   */
  def stop(): Unit =
    running.synchronized {
      if (running.get) {
        running.getAndSet(false)
        executorService.shutdown()
        executorService.awaitTermination(5000, TimeUnit.MILLISECONDS)
        client.shutdown()
      }
    }
}

/**
 * Runnable task for asynchronously submitting metrics to CloudWatch
 * @param client CW client
 * @param name name of the metric to emit
 * @param value numeric value for the metric
 * @param tags key-value pairs to associate with the metric
 */
class EmitMetricTask(
    client: AmazonCloudWatch,
    name: String,
    value: Double,
    tags: Map[String, String]
) extends Runnable
    with LazyLogging {
  override def run(): Unit = {
    val datum = new MetricDatum()
      .withMetricName(name)
      .withTimestamp(new Date())
      .withUnit(StandardUnit.Count)
      .withValue(value)

    if (tags.nonEmpty) {
      datum.setDimensions(convertMapToDimensions(tags).asJava)
    }

    logger.info("Emitting new metric[{}={}]", name, value)
    val request = new PutMetricDataRequest()
      .withNamespace("TwitchSongRequests")
      .withMetricData(datum)

    val response = client.putMetricData(request)
    logger.info(
        "Submitted metric data point for {}={}. Responded with HTTP status code {} and request ID {}",
        name,
        value,
        response.getSdkHttpMetadata.getHttpStatusCode,
        response.getSdkResponseMetadata.getRequestId
    )
  }

  /**
   * Utility method that transforms the key-value pairs into the required
   * CloudWatch POJOs
   * @param tags key-value pairs for the metric
   * @return list of Dimension objects based on the key-value pairs
   */
  private def convertMapToDimensions(
      tags: Map[String, String]
  ): Seq[Dimension] =
    tags.map { case (k, v) =>
      new Dimension().withName(k).withValue(v)
    }.toSeq
}
