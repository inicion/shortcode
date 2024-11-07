#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Check if CDK is installed
if ! command -v cdk &> /dev/null; then
    echo "CDK is not installed. Please install AWS CDK (npm install -g aws-cdk) and try again."
    exit 1
fi

# Default values for tableName and domainName
DEFAULT_TABLE_NAME="ShortCodes"
DEFAULT_DOMAIN_NAME="o-sp.one"

# Use arguments if provided, otherwise use defaults
TABLE_NAME=${1:-$DEFAULT_TABLE_NAME}
DOMAIN_NAME=${2:-$DEFAULT_DOMAIN_NAME}

# Compile Go binary for Linux (required for AWS Lambda)
echo "Building Go binary for Lambda..."
cd src/url-shortener
GOOS=linux GOARCH=amd64 go build -o ../../bin/main main.go
cd ../..

# Package the binary for Lambda deployment
echo "Packaging Go binary..."
cd bin
zip -j main.zip main
cd ..

# Install Node dependencies if not present
if [ ! -d "node_modules" ]; then
    echo "Installing Node dependencies..."
    npm install
fi

npm run build
# Bootstrap CDK (only needed once per environment)
if ! cdk bootstrap; then
    echo "CDK bootstrap failed. Ensure your AWS credentials are configured correctly and try again."
    exit 1
fi

# Synthesize the CloudFormation template
echo "Synthesizing the CloudFormation template..."
if ! cdk synth; then
    echo "CDK synthesis failed. Check your TypeScript code for errors and try again."
    exit 1
fi

# Deploy the stack with tableName and domainName parameters
echo "Deploying the stack..."
if ! cdk deploy --require-approval never --context tableName=$TABLE_NAME --context domainName=$DOMAIN_NAME; then
    echo "CDK deployment failed. Check the error message above for details."
    exit 1
fi

echo "Deployment successful!"
