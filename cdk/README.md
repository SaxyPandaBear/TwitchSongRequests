# Getting Started
1. `npm install -g aws-cdk`
2. `npm install -g typescript`
3. `npm install @aws-cdk/core @aws-cdk/aws-lambda @aws-cdk/aws-lambda-event-sources @aws-cdk/aws-sqs`
4. Ensure that `npm run build` and `cdk synth` run successfully
5. If adding new dependencies to the CDK stack first run `npm install <DEPENDENCY>`
6. Always test changes by making sure `npm run build` and `cdk synth` run successfully and verifying that the output matches what is expected


The `cdk.json` file tells the CDK Toolkit how to execute your app.

## Useful commands

 * `npm run build`   compile typescript to js
 * `npm run watch`   watch for changes and compile
 * `cdk deploy`      deploy this stack to your default AWS account/region
 * `cdk diff`        compare deployed stack with current state
 * `cdk synth`       emits the synthesized CloudFormation template
