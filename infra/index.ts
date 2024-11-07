#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { URLShortenerStack } from './url-shortener-stack';

const app = new cdk.App();
console.log(app.node.tryGetContext("tableName"));
console.log(app.node.tryGetContext("domainName"));
console.log(app.node.tryGetContext("accountId"));
console.log(app.node.tryGetContext("region"));



new URLShortenerStack(app, 'URLShortenerStack', {
    tableName: app.node.tryGetContext("tableName"),
    domainName: app.node.tryGetContext("domainName"),
    env: {
        account: app.node.tryGetContext("accountId"),
        region: app.node.tryGetContext("region"),
    },
});