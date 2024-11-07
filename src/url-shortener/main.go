package main

import (
	"context"
	"log"
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
	region    = os.Getenv("REGION")
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}
	dbClient = dynamodb.NewFromConfig(cfg)
	log.Println("DynamoDB client initialized")
}

// Handler for the Lambda function
func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Handling request")

	// Retrieve the shortcode from the path parameters
	shortcode, exists := request.PathParameters["code"]
	if !exists || shortcode == "" {
		log.Println("Missing 'code' parameter")
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
			"SortKey":   &types.AttributeValueMemberS{Value: "META"},
		},
	})
	if err != nil || result.Item == nil {
		log.Printf("Shortcode not found: %v", err)
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
			"SortKey":   &types.AttributeValueMemberS{Value: "VIEW#" + time.Now().Format(time.RFC3339)},
			"Timestamp": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
			"Action":    &types.AttributeValueMemberS{Value: "redirect"},
		},
	})
	if err != nil {
		log.Printf("Error logging usage: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Error logging usage",
		}, nil
	}

	// Return a redirect response
	log.Printf("Redirecting to: %s", longURL)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusFound,
		Headers: map[string]string{
			"Location": longURL,
		},
	}, nil
}

func main() {
	log.Println("Starting Lambda function")
	lambda.Start(handleRequest)
}
