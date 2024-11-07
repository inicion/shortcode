#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { URLShortenerStack } from './url-shortener-stack';

const app = new cdk.App();
new URLShortenerStack(app, 'URLShortenerStack', {
    tableName: process.env.DYNAMODB_TABLE_NAME || 'DefaultTableName',
    domainName: process.env.DOMAIN_NAME || 'example.com',
    env: {
        region: process.env.AWS_REGION || 'us-east-1',
    },
});