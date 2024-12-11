import * as cdk from 'aws-cdk-lib';
import * as cognito from 'aws-cdk-lib/aws-cognito';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as cloudfront from 'aws-cdk-lib/aws-cloudfront';
import * as cloudfrontOrigins from 'aws-cdk-lib/aws-cloudfront-origins';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as route53 from 'aws-cdk-lib/aws-route53';
import * as targets from 'aws-cdk-lib/aws-route53-targets';
import * as acm from 'aws-cdk-lib/aws-certificatemanager';
import * as path from 'path';

export class URLShortenerStack extends cdk.Stack {
    constructor(scope: cdk.App, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const tableName = this.node.tryGetContext('tableName') || props.tableName;
        const domainName = this.node.tryGetContext('domainName') || props.domainName;
        const region = this.node.tryGetContext('region') || props.env?.region;
        const subdomain = 'app';

        // Cognito User Pool
        const userPool = new cognito.UserPool(this, 'UserPool', {
            selfSignUpEnabled: true,
            signInAliases: { email: true },
            autoVerify: { email: true },
        });

        const userPoolClient = new cognito.UserPoolClient(this, 'UserPoolClient', {
            userPool,
        });

        // S3 Bucket for Static Site
        const siteBucket = new s3.Bucket(this, 'SiteBucket', {
            websiteIndexDocument: 'index.html',
            publicReadAccess: true,
            removalPolicy: cdk.RemovalPolicy.DESTROY,
        });

        // CloudFront Distribution for Static Site
        const distribution = new cloudfront.Distribution(this, 'SiteDistribution', {
            defaultBehavior: { origin: new cloudfrontOrigins.S3Origin(siteBucket) },
            domainNames: [`${subdomain}.${domainName}`],
            certificate: new acm.Certificate(this, 'SiteCertificate', {
                domainName: `${subdomain}.${domainName}`,
                validation: acm.CertificateValidation.fromDns(),
            }),
        });

        // Route 53 Alias Record for CloudFront
        const hostedZone = route53.HostedZone.fromLookup(this, 'HostedZone', {
            domainName,
        });

        new route53.ARecord(this, 'SiteAliasRecord', {
            zone: hostedZone,
            target: route53.RecordTarget.fromAlias(new targets.CloudFrontTarget(distribution)),
            recordName: subdomain,
        });

        // DynamoDB Table
        const table = new dynamodb.Table(this, 'ShortcodesTable', {
            tableName: tableName,
            partitionKey: { name: 'Shortcode', type: dynamodb.AttributeType.STRING },
            sortKey: { name: 'SortKey', type: dynamodb.AttributeType.STRING },
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
        });

        // Lambda Function
        const shortcodesLambda = new lambda.Function(this, 'URLShortenerHandler', {
            runtime: lambda.Runtime.PROVIDED_AL2,
            handler: 'main',
            code: lambda.Code.fromAsset(path.join(__dirname, '../bin/main.zip')),
            architecture: lambda.Architecture.ARM_64,
            environment: {
                DYNAMODB_TABLE_NAME: table.tableName,
                REGION: region,
                USER_POOL_ID: userPool.userPoolId,
                USER_POOL_CLIENT_ID: userPoolClient.userPoolClientId,
            },
        });

        table.grantReadWriteData(shortcodesLambda);

        // API Gateway
        const api = new apigateway.RestApi(this, 'ShortcodesApi', {
            restApiName: 'Shortcodes Service',
            description: 'This service handles shortcodes.',
        });

        const authorizer = new apigateway.CognitoUserPoolsAuthorizer(this, 'Authorizer', {
            cognitoUserPools: [userPool],
        });

        const shortcodes = api.root.addResource('shortcodes');
        shortcodes.addMethod('GET', new apigateway.LambdaIntegration(shortcodesLambda), {
            authorizer,
            authorizationType: apigateway.AuthorizationType.COGNITO,
        });
        shortcodes.addMethod('POST', new apigateway.LambdaIntegration(shortcodesLambda), {
            authorizer,
            authorizationType: apigateway.AuthorizationType.COGNITO,
        });

        const shortcode = shortcodes.addResource('{code}');
        shortcode.addMethod('GET', new apigateway.LambdaIntegration(shortcodesLambda), {
            authorizer,
            authorizationType: apigateway.AuthorizationType.COGNITO,
        });
        shortcode.addMethod('PUT', new apigateway.LambdaIntegration(shortcodesLambda), {
            authorizer,
            authorizationType: apigateway.AuthorizationType.COGNITO,
        });
        shortcode.addMethod('DELETE', new apigateway.LambdaIntegration(shortcodesLambda), {
            authorizer,
            authorizationType: apigateway.AuthorizationType.COGNITO,
        });

        const metrics = api.root.addResource('metrics');
        metrics.addMethod('GET', new apigateway.LambdaIntegration(shortcodesLambda), {
            authorizer,
            authorizationType: apigateway.AuthorizationType.COGNITO,
        });

        const metric = metrics.addResource('{code}');
        metric.addMethod('GET', new apigateway.LambdaIntegration(shortcodesLambda), {
            authorizer,
            authorizationType: apigateway.AuthorizationType.COGNITO,
        });

        // Public Access to Shortcodes
        const publicShortcodes = api.root.addResource('s');
        const publicShortcode = publicShortcodes.addResource('{code}');
        publicShortcode.addMethod('GET', new apigateway.LambdaIntegration(shortcodesLambda));
    }
}