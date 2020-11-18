package com.github.saxypandabear.songrequests.lib

import com.amazonaws.regions.Region
import com.amazonaws.services.cloudwatch.AmazonCloudWatch
import com.amazonaws.services.cloudwatch.model._
import com.amazonaws.services.cloudwatch.waiters.AmazonCloudWatchWaiters
import com.amazonaws.{AmazonWebServiceRequest, ResponseMetadata}

import scala.collection.mutable

//noinspection ScalaStyle
class DummyAmazonCloudWatch extends AmazonCloudWatch {
  val putMetricDataRequests = new mutable.ArrayBuffer[PutMetricDataRequest]()

  override def setEndpoint(endpoint: String): Unit =
    throw new UnsupportedOperationException

  override def setRegion(region: Region): Unit =
    throw new UnsupportedOperationException

  override def deleteAlarms(
      deleteAlarmsRequest: DeleteAlarmsRequest
  ): DeleteAlarmsResult =
    throw new UnsupportedOperationException

  override def deleteAnomalyDetector(
      deleteAnomalyDetectorRequest: DeleteAnomalyDetectorRequest
  ): DeleteAnomalyDetectorResult =
    throw new UnsupportedOperationException

  override def deleteDashboards(
      deleteDashboardsRequest: DeleteDashboardsRequest
  ): DeleteDashboardsResult = throw new UnsupportedOperationException

  override def deleteInsightRules(
      deleteInsightRulesRequest: DeleteInsightRulesRequest
  ): DeleteInsightRulesResult = throw new UnsupportedOperationException

  override def describeAlarmHistory(
      describeAlarmHistoryRequest: DescribeAlarmHistoryRequest
  ): DescribeAlarmHistoryResult = throw new UnsupportedOperationException

  override def describeAlarmHistory(): DescribeAlarmHistoryResult =
    throw new UnsupportedOperationException

  override def describeAlarms(
      describeAlarmsRequest: DescribeAlarmsRequest
  ): DescribeAlarmsResult = throw new UnsupportedOperationException

  override def describeAlarms(): DescribeAlarmsResult =
    throw new UnsupportedOperationException

  override def describeAlarmsForMetric(
      describeAlarmsForMetricRequest: DescribeAlarmsForMetricRequest
  ): DescribeAlarmsForMetricResult = throw new UnsupportedOperationException

  override def describeAnomalyDetectors(
      describeAnomalyDetectorsRequest: DescribeAnomalyDetectorsRequest
  ): DescribeAnomalyDetectorsResult = throw new UnsupportedOperationException

  override def describeInsightRules(
      describeInsightRulesRequest: DescribeInsightRulesRequest
  ): DescribeInsightRulesResult = throw new UnsupportedOperationException

  override def disableAlarmActions(
      disableAlarmActionsRequest: DisableAlarmActionsRequest
  ): DisableAlarmActionsResult = throw new UnsupportedOperationException

  override def disableInsightRules(
      disableInsightRulesRequest: DisableInsightRulesRequest
  ): DisableInsightRulesResult = throw new UnsupportedOperationException

  override def enableAlarmActions(
      enableAlarmActionsRequest: EnableAlarmActionsRequest
  ): EnableAlarmActionsResult = throw new UnsupportedOperationException

  override def enableInsightRules(
      enableInsightRulesRequest: EnableInsightRulesRequest
  ): EnableInsightRulesResult = throw new UnsupportedOperationException

  override def getDashboard(
      getDashboardRequest: GetDashboardRequest
  ): GetDashboardResult = throw new UnsupportedOperationException

  override def getInsightRuleReport(
      getInsightRuleReportRequest: GetInsightRuleReportRequest
  ): GetInsightRuleReportResult = throw new UnsupportedOperationException

  override def getMetricData(
      getMetricDataRequest: GetMetricDataRequest
  ): GetMetricDataResult = throw new UnsupportedOperationException

  override def getMetricStatistics(
      getMetricStatisticsRequest: GetMetricStatisticsRequest
  ): GetMetricStatisticsResult = throw new UnsupportedOperationException

  override def getMetricWidgetImage(
      getMetricWidgetImageRequest: GetMetricWidgetImageRequest
  ): GetMetricWidgetImageResult = throw new UnsupportedOperationException

  override def listDashboards(
      listDashboardsRequest: ListDashboardsRequest
  ): ListDashboardsResult = throw new UnsupportedOperationException

  override def listMetrics(
      listMetricsRequest: ListMetricsRequest
  ): ListMetricsResult = throw new UnsupportedOperationException

  override def listMetrics(): ListMetricsResult =
    throw new UnsupportedOperationException

  override def listTagsForResource(
      listTagsForResourceRequest: ListTagsForResourceRequest
  ): ListTagsForResourceResult = throw new UnsupportedOperationException

  override def putAnomalyDetector(
      putAnomalyDetectorRequest: PutAnomalyDetectorRequest
  ): PutAnomalyDetectorResult = throw new UnsupportedOperationException

  override def putCompositeAlarm(
      putCompositeAlarmRequest: PutCompositeAlarmRequest
  ): PutCompositeAlarmResult = throw new UnsupportedOperationException

  override def putDashboard(
      putDashboardRequest: PutDashboardRequest
  ): PutDashboardResult = throw new UnsupportedOperationException

  override def putInsightRule(
      putInsightRuleRequest: PutInsightRuleRequest
  ): PutInsightRuleResult = throw new UnsupportedOperationException

  override def putMetricAlarm(
      putMetricAlarmRequest: PutMetricAlarmRequest
  ): PutMetricAlarmResult = throw new UnsupportedOperationException

  override def putMetricData(
      putMetricDataRequest: PutMetricDataRequest
  ): PutMetricDataResult = {
    putMetricDataRequests += putMetricDataRequest
    new PutMetricDataResult()
  }

  override def setAlarmState(
      setAlarmStateRequest: SetAlarmStateRequest
  ): SetAlarmStateResult = throw new UnsupportedOperationException

  override def tagResource(
      tagResourceRequest: TagResourceRequest
  ): TagResourceResult = throw new UnsupportedOperationException

  override def untagResource(
      untagResourceRequest: UntagResourceRequest
  ): UntagResourceResult = throw new UnsupportedOperationException

  override def shutdown(): Unit = throw new UnsupportedOperationException

  override def getCachedResponseMetadata(
      request: AmazonWebServiceRequest
  ): ResponseMetadata = throw new UnsupportedOperationException

  override def waiters(): AmazonCloudWatchWaiters =
    throw new UnsupportedOperationException
}
