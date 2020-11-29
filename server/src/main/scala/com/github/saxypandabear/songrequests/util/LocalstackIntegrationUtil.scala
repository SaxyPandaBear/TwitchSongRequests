package com.github.saxypandabear.songrequests.util

object LocalstackIntegrationUtil {
  // From the GitHub Actions docs:
  // Always set to true when GitHub Actions is running the workflow. You can use
  // this variable to differentiate when tests are being run locally or by
  // GitHub Actions.
  private val GITHUB_ACTIONS = "GITHUB_ACTIONS"

  // this should match the name of the service defined in the GitHub workflow
  // actions
  private val LOCALSTACK_HOSTNAME = "localstack"
  private val LOCALHOST           = "localhost"

  /**
   * Because localstack defaults to `localhost`, this does not work when
   * running in the CI workflow. This is because the CI workflow stands up a
   * container with networking mapping the localstack image to "localstack",
   * rather than running on pure localhost, like it does locally.
   *
   * Use a system environment variable to help determine if this is being run
   * in the remote CI workflow.
   * https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables
   * GitHub reserves environment variable names with the "GITHUB_" prefix, and
   * I don't think we will ever need to configure an environment variable with
   * that prefix, so the existence of one of those variable names in the system
   * should be sufficient.
   * @param localstackUrl original localstack URL for the service
   * @return localstackUrl if no GitHub env vars are present, otherwise the
   *         input URL with "localhost" replaced with "localstack"
   */
  def resolveCorrectUrl(localstackUrl: String): String = {
    val projectProperties = new ProjectProperties().withSystemProperties()
    if (projectProperties.has(GITHUB_ACTIONS)) {
      localstackUrl.replace(LOCALHOST, LOCALSTACK_HOSTNAME)
    } else {
      localstackUrl
    }
  }
}
