package com.github.saxypandabear.songrequests.metric

import java.util.Date

import com.amazonaws.services.cloudwatch.AmazonCloudWatch
import com.amazonaws.services.cloudwatch.model.{Dimension, MetricDatum, PutMetricDataRequest, StandardUnit}
import com.typesafe.scalalogging.LazyLogging

import scala.collection.JavaConverters._

/**
 * Simple implementation that takes data and publishes to CloudWatch
 * @param client internal AWS SDK CloudWatch client
 */
class CloudWatchMetricCollector(client: AmazonCloudWatch) extends LazyLogging {
    def emitCountMetric(name: String, value: Double, tags: Map[String, String] = Map.empty): Unit = {
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
        logger.info("Response from emitting metric[{}={}]: {}", name, value, response)
    }

    private def convertMapToDimensions(tags: Map[String, String]): Seq[Dimension] = {
        tags.map {
            case (k, v) => new Dimension().withName(k).withValue(v)
        }.toSeq
    }
}
