package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
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

func generateShortcode(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// Handler for the Lambda function
func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Handling request")

	switch request.Resource {
	case "/generate":
		return handleGenerate(ctx, request)
	case "/s/{code}":
		return handleRedirect(ctx, request)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "Not Found",
		}, nil
	}
}

func handleGenerate(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var body struct {
		URL         string `json:"url"`
		Shortcode   string `json:"shortcode,omitempty"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(request.Body), &body); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid request body",
		}, nil
	}

	if body.URL == "" || body.Description == "" {
		log.Println("Missing 'url' or 'description' in request body")
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Missing 'url' or 'description' in request body",
		}, nil
	}

	shortcode := body.Shortcode
	if shortcode == "" {
		for i := 0; i < 5; i++ { // Try up to 5 times to generate a unique shortcode
			shortcode = generateShortcode(4)
			// Check if the shortcode already exists
			result, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
				TableName: aws.String(tableName),
				Key: map[string]types.AttributeValue{
					"Shortcode": &types.AttributeValueMemberS{Value: shortcode},
					"SortKey":   &types.AttributeValueMemberS{Value: "META"},
				},
			})
			if err != nil {
				log.Printf("Error checking shortcode: %v", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       "Error checking shortcode",
				}, nil
			}
			if result.Item == nil {
				break // Shortcode is unique
			}
			if i == 4 {
				log.Println("Failed to generate a unique shortcode after 5 attempts")
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       "Failed to generate a unique shortcode",
				}, nil
			}
		}
	}

	_, err := dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"Shortcode":   &types.AttributeValueMemberS{Value: shortcode},
			"SortKey":     &types.AttributeValueMemberS{Value: "META"},
			"URL":         &types.AttributeValueMemberS{Value: body.URL},
			"Description": &types.AttributeValueMemberS{Value: body.Description},
		},
	})
	if err != nil {
		log.Printf("Error saving shortcode: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Error saving shortcode",
		}, nil
	}

	responseBody, _ := json.Marshal(map[string]string{
		"shortcode": shortcode,
	})
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(responseBody),
	}, nil
}

func handleRedirect(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	shortcode, exists := request.PathParameters["code"]
	if !exists || shortcode == "" {
		log.Println("Missing 'code' parameter")
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Missing 'code' parameter",
		}, nil
	}

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

	longURL := result.Item["URL"].(*types.AttributeValueMemberS).Value

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
