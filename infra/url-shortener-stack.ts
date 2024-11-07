import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as route53 from 'aws-cdk-lib/aws-route53';
import * as targets from 'aws-cdk-lib/aws-route53-targets';
import * as acm from 'aws-cdk-lib/aws-certificatemanager';
// import * as iam from 'aws-cdk-lib/aws-iam';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as path from 'path';

interface URLShortenerStackProps extends cdk.StackProps {
    tableName: string;
    domainName: string;
}

export class URLShortenerStack extends cdk.Stack {
    constructor(scope: cdk.App, id: string, props: URLShortenerStackProps) {
        super(scope, id, props);

        const tableName = this.node.tryGetContext('tableName') || props.tableName;
        const domainName = this.node.tryGetContext('domainName') || props.domainName;
        const region = this.node.tryGetContext('region') || props.env?.region;

        // DynamoDB Table
        const table = new dynamodb.Table(this, 'ShortcodesTable', {
            tableName: tableName,
            partitionKey: { name: 'Shortcode', type: dynamodb.AttributeType.STRING },
            sortKey: { name: 'SortKey', type: dynamodb.AttributeType.STRING },
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
        });

        // Global Secondary Index for querying view logs
        table.addGlobalSecondaryIndex({
            indexName: 'ShortcodeIndex',
            partitionKey: { name: 'Shortcode', type: dynamodb.AttributeType.STRING },
            sortKey: { name: 'Timestamp', type: dynamodb.AttributeType.STRING },
            projectionType: dynamodb.ProjectionType.ALL,
        });
        // Lambda Function
        const urlShortenerLambda = new lambda.Function(this, 'URLShortenerHandler', {
            runtime: lambda.Runtime.PROVIDED_AL2,
            handler: 'main',
            code: lambda.Code.fromAsset(path.join(__dirname, '../bin/main.zip')),
            architecture: lambda.Architecture.ARM_64,
            environment: {
                DYNAMODB_TABLE_NAME: table.tableName,
                REGION: region,
            },
        });

        // Grant the Lambda function permissions to access DynamoDB
        table.grantReadWriteData(urlShortenerLambda);
        // API Gateway
        const api = new apigateway.LambdaRestApi(this, 'UrlShortenerApi', {
            handler: urlShortenerLambda,
            proxy: false,
        });

        api.root.addResource('{code}').addMethod('GET');

        // Add a default ANY method to the root resource
        api.root.addMethod('ANY');

        // Route 53 Hosted Zone Lookup
        const hostedZone = route53.HostedZone.fromLookup(this, 'HostedZone', {
            domainName: domainName,
        });

        // Request an SSL Certificate for the custom domain
        const certificate = new acm.Certificate(this, 'ApiGatewayCertificate', {
            domainName: domainName, // Use the domain name for HTTPS
            validation: acm.CertificateValidation.fromDns(hostedZone),
        });

        // Create a custom domain for API Gateway
        const customDomain = new apigateway.DomainName(this, 'CustomDomain', {
            domainName: domainName,
            certificate: certificate,
            endpointType: apigateway.EndpointType.REGIONAL,
        });

        // Map the custom domain to the API Gateway stage
        customDomain.addBasePathMapping(api, { basePath: '' });

        // Route 53 Alias Record for the custom domain
        new route53.ARecord(this, 'ApiGatewayAliasRecord', {
            zone: hostedZone,
            target: route53.RecordTarget.fromAlias(new targets.ApiGatewayDomain(customDomain)),
        });
    }
}