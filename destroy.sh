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


# Confirm with the user
read -p "Are you sure you want to destroy the stack? This action cannot be undone. (y/N): " confirm
if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
    echo "Destroy canceled."
    exit 0
fi

# Destroy the stack
echo "Destroying the stack..."
if ! cdk destroy -c accountId=$AWS_ACCOUNT \
    -c tableName=$TABLE_NAME \
    -c region=$AWS_REGION \
    -c domainName=$DOMAIN_NAME --force; then
    echo "CDK destroy failed. Check the error message above for details."
    exit 1
fi

# Clean up packaged Lambda binary
rm -f src/url-shortener/main.zip

echo "Stack destroyed and package cleanup complete!"
