#!/bin/bash

# Check if CDK is installed
if ! command -v cdk &> /dev/null; then
    echo "CDK is not installed. Please install AWS CDK (npm install -g aws-cdk) and try again."
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
if ! cdk destroy --force; then
    echo "CDK destroy failed. Check the error message above for details."
    exit 1
fi

# Clean up packaged Lambda binary
rm -f src/url-shortener/main.zip

echo "Stack destroyed and package cleanup complete!"
