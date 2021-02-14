package com.github.saxypandabear.songrequests.util

import com.amazonaws.client.builder.AwsClientBuilder.EndpointConfiguration
import com.amazonaws.client.builder.AwsSyncClientBuilder
import com.amazonaws.services.cloudwatch.{
  AmazonCloudWatch,
  AmazonCloudWatchClientBuilder
}
import com.amazonaws.services.dynamodbv2.{
  AmazonDynamoDB,
  AmazonDynamoDBClientBuilder
}
import com.amazonaws.services.sqs.{AmazonSQS, AmazonSQSClientBuilder}

object AwsUtil {

  def createCloudWatchClient(
      projectProperties: ProjectProperties
  ): AmazonCloudWatch = {
    val cloudWatchBuilder = AmazonCloudWatchClientBuilder.standard()
    setLocalStackUrlIfPresentElseRegion[
        AmazonCloudWatchClientBuilder,
        AmazonCloudWatch,
        AmazonCloudWatchClientBuilder
    ](
        cloudWatchBuilder,
        "cloudwatch.url",
        projectProperties
    ).build()
  }

  def createSqsClient(
      projectProperties: ProjectProperties
  ): AmazonSQS = {
    val sqsBuilder = AmazonSQSClientBuilder.standard()
    setLocalStackUrlIfPresentElseRegion[
        AmazonSQSClientBuilder,
        AmazonSQS,
        AmazonSQSClientBuilder
    ](
        sqsBuilder,
        "sqs.url",
        projectProperties
    ).build()
  }

  def createDynamoDbClient(
      projectProperties: ProjectProperties
  ): AmazonDynamoDB = {
    val dynamoDbBuilder = AmazonDynamoDBClientBuilder.standard()
    setLocalStackUrlIfPresentElseRegion[
        AmazonDynamoDBClientBuilder,
        AmazonDynamoDB,
        AmazonDynamoDBClientBuilder
    ](
        dynamoDbBuilder,
        "dynamodb.url",
        projectProperties
    ).build()
  }

  /**
   * The problem is that in order to properly interact with Localstack, we need
   * to set the URL to the specified localstack URL.
   * This expects a "region" parameter in the properties object, else it
   * defaults to 'us-east-1' for the region
   * @param builder the AWS client builder object
   * @param key the key to search for in the properties object
   * @param projectProperties enumerates all of the application and system properties for this
   * @tparam Builder the Builder type
   * @tparam Type the AWS Service
   * @tparam T The specific AWS Service Builder type
   * @return the builder that was passed in, with either an endpoint
   *         configuration or region configuration
   */
  private def setLocalStackUrlIfPresentElseRegion[
      Builder <: T,
      Type,
      T <: AwsSyncClientBuilder[Builder, Type]
  ](
      builder: T,
      key: String,
      projectProperties: ProjectProperties
  ): T = {
    val region = projectProperties.getString("region").getOrElse("us-east-1")
    projectProperties.getString(key) match {
      case Some(url) =>
        val endpointConfiguration = new EndpointConfiguration(url, region)
        builder.setEndpointConfiguration(endpointConfiguration)
      case None      => builder.setRegion(region)
    }

    builder
  }
}
