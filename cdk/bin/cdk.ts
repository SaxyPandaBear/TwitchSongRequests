#!/usr/bin/env node
import 'source-map-support/register';
import {App} from 'monocdk';
import {CdkStack} from '../lib/cdk-stack';

const app = new App();
new CdkStack(app, 'CdkStack');
