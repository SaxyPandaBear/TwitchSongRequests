package com.github.saxypandabear.songrequests.intercept
import com.amazonaws.services.sqs.AmazonSQS
import com.github.saxypandabear.songrequests.intercept.model.Message
import com.github.saxypandabear.songrequests.metric.CloudWatchMetricCollector
import com.typesafe.scalalogging.LazyLogging

class SQSEventInterceptor(
    sqs: AmazonSQS,
    metricsCollector: CloudWatchMetricCollector
) extends EventInterceptor
    with LazyLogging {
  override def poll(): Seq[Message] = ???

  override def shutdown(): Unit = {
    sqs.shutdown()
    metricsCollector.stop()
  }
}
