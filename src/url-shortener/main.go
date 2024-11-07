package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	dbClient  *dynamodb.Client
	tableName = os.Getenv("DYNAMODB_TABLE_NAME")
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-2"))
	if err != nil {
		panic("unable to load AWS config")
	}
	dbClient = dynamodb.NewFromConfig(cfg)
}

// Handler for the Lambda function
func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Retrieve the shortcode from the path parameters
	shortcode, exists := request.PathParameters["code"]
	if !exists || shortcode == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Missing 'code' parameter",
		}, nil
	}

	// Fetch the URL associated with the shortcode from DynamoDB
	result, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"Shortcode": &types.AttributeValueMemberS{Value: shortcode},
		},
	})
	if err != nil || result.Item == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "Shortcode not found",
		}, nil
	}

	// Retrieve the long URL from the result
	longURL := result.Item["URL"].(*types.AttributeValueMemberS).Value

	// Log the usage to DynamoDB
	_, err = dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"Shortcode": &types.AttributeValueMemberS{Value: shortcode},
			"Timestamp": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
			"Action":    &types.AttributeValueMemberS{Value: "redirect"},
		},
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Error logging usage",
		}, nil
	}

	// Return a redirect response
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusFound,
		Headers: map[string]string{
			"Location": longURL,
		},
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
