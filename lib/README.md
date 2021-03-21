Shared Libs
===========

This package contains all the shared code that is reused across the
different components. This includes things like the DynamoDB 
encapsulation, CloudWatch metrics collection, various utility classes,
and test fixtures that can be reused in other Scala packages in the 
mono-repo for a reusable module for testing code.

To include the library module, add this to the `dependencies`
closure of the `build.gradle` file:
```groovy
implementation project(':lib')
```

To include the library test fixtures, add this to the `dependencies`
closure of the `build.gradle` file:
```groovy
testImplementation(testFixtures(project(':lib')))
```
Also, make sure that Scala test classes are built before Java test classes:
```groovy
sourceSets.test.scala.srcDir "src/test/java"
sourceSets.test.java.srcDirs = []
```
This is necessary because there is a dependency on Scala source code
when writing tests for the Scala classes, in Java. This is most 
noticeable in things like Localstack integration tests, since those
must be written in Java.

### Build
Run `gradle :lib:assemble` to build the module

Run `gradle :lib:scalaDocs` to generate ScalaDocs for the module

### Test
`gradle :lib:test`

### Run locally
Not applicable, since this is a shared library module.
