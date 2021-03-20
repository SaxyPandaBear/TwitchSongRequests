# Getting Started
1. Navigate to the cdk directory in terminal
2. `npm install monocdk`
3. Run `npm run build` and `cdk synth` to test that everything is setup correctly and test cdk changes


# CDK References
* [AWS CDK API Reference](https://docs.aws.amazon.com/cdk/api/latest/docs/aws-construct-library.html)
* [AWS CDK examples using TypeScript](https://github.com/aws-samples/aws-cdk-examples/tree/master/typescript)
* [AWS intro to CDK](https://docs.aws.amazon.com/cdk/latest/guide/hello_world.html)

The `cdk.json` file tells the CDK Toolkit how to execute your app.

## Useful commands

 * `npm run build`   compile typescript to js
 * `npm run watch`   watch for changes and compile
 * `cdk deploy`      deploy this stack to your default AWS account/region
 * `cdk diff`        compare deployed stack with current state
 * `cdk synth`       emits the synthesized CloudFormation template
