import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as route53 from 'aws-cdk-lib/aws-route53';
import * as targets from 'aws-cdk-lib/aws-route53-targets';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as path from 'path';

interface URLShortenerStackProps extends cdk.StackProps {
    tableName: string;
    domainName: string;
}

export class URLShortenerStack extends cdk.Stack {
    constructor(scope: cdk.App, id: string, props: URLShortenerStackProps) {
        super(scope, id, props);

        // Use context values for table name and domain name if passed through deploy.sh
        const tableName = this.node.tryGetContext('tableName') || props.tableName;
        const domainName = this.node.tryGetContext('domainName') || props.domainName;

        // DynamoDB Table
        const table = new dynamodb.Table(this, 'ShortcodesTable', {
            tableName: tableName,
            partitionKey: { name: 'Shortcode', type: dynamodb.AttributeType.STRING },
            sortKey: { name: 'Timestamp', type: dynamodb.AttributeType.STRING },
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
        });

        // Lambda Function
        const urlShortenerLambda = new lambda.Function(this, 'URLShortenerHandler', {
            runtime: lambda.Runtime.PROVIDED_AL2,
            handler: 'main',
            code: lambda.Code.fromAsset(path.join(__dirname, '../bin/main.zip')),
            environment: {
                DYNAMODB_TABLE_NAME: table.tableName,
            },
        });

        // Grant the Lambda function permissions to access DynamoDB
        table.grantReadWriteData(urlShortenerLambda);

        // API Gateway Setup
        const api = new apigateway.LambdaRestApi(this, 'UrlShortenerApi', {
            handler: urlShortenerLambda,
            proxy: false,
        });

        const shortcodeResource = api.root.addResource('s');
        shortcodeResource.addResource('{code}').addMethod('GET');

        // Route 53 Setup
        const hostedZone = route53.HostedZone.fromLookup(this, 'HostedZone', {
            domainName: domainName,
        });

        new route53.ARecord(this, 'ApiGatewayAliasRecord', {
            zone: hostedZone,
            target: route53.RecordTarget.fromAlias(new targets.ApiGateway(api)),
            recordName: 'short', // Replace with your desired subdomain
        });
    }
}
