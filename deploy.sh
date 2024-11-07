#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

source .env

# Check if CDK is installed
if ! command -v cdk &> /dev/null; then
    echo "CDK is not installed. Please install AWS CDK (npm install -g aws-cdk) and try again."
    exit 1
fi

# Check if all required environment variables are set
if [ -z "$AWS_ACCOUNT" ] || [ -z "$AWS_REGION" ] || [ -z "$TABLE_NAME" ] || [ -z "$DOMAIN_NAME" ]; then
    echo "One or more required environment variables are not set."
    echo "Please set the following environment variables and try again:"
    echo -e "\tAWS_ACCOUNT"
    echo -e "\tAWS_REGION"
    echo -e "\tTABLE_NAME"
    echo -e "\tDOMAIN_NAME"
    exit 1
fi

# Compile Go binary for Linux (required for AWS Lambda)
echo "Building Go binary for Lambda..."
cd src/url-shortener
# GOOS=linux GOARCH=amd64 go build -o main main.go
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ../../bin/bootstrap main.go
# mv bootstrap ../../bin/bootstrap

cd ../..

# Package the binary for Lambda deployment
echo "Packaging Go binary..."
cd bin
zip -j main.zip bootstrap
cd ..

# Install Node dependencies if not present
if [ ! -d "node_modules" ]; then
    echo "Installing Node dependencies..."
    npm install
fi

npm run build

if ! cdk deploy -c accountId=$AWS_ACCOUNT \
    -c tableName=$TABLE_NAME \
    -c region=$AWS_REGION \
    -c domainName=$DOMAIN_NAME; then
    echo "Deployment failed. Bootstrapping environment..."
    cdk bootstrap aws://$AWS_ACCOUNT/$AWS_REGION \
        -c region=$AWS_REGION \
        -c tableName=$TABLE_NAME \
        -c accountId=$AWS_ACCOUNT \
        -c domainName=$AWS_ACCOUNT \
        --cloudformation-execution-policies arn:aws:iam::aws:policy/AdministratorAccess
fi
echo "Deployment successful!"
