package utils

import (
	"context"
	"log"
	"os"
	"strings"
	"time"
	"url-shortener/models"

	"github.com/aws/aws-lambda-go/events"
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

func GetDynamoDBItem(ctx context.Context, shortcode string) (map[string]types.AttributeValue, error) {
	result, err := dbClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"Shortcode": &types.AttributeValueMemberS{Value: shortcode},
			"SortKey":   &types.AttributeValueMemberS{Value: "META"},
		},
	})
	if err != nil {
		return nil, err
	}
	return result.Item, nil
}

func PutDynamoDBItem(ctx context.Context, item map[string]types.AttributeValue) error {
	_, err := dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	return err
}

func LogRedirect(ctx context.Context, shortcode string, request events.APIGatewayProxyRequest) error {
	// Extract platform information from User-Agent header
	userAgent := request.Headers["User-Agent"]
	platform := "unknown"
	if strings.Contains(strings.ToLower(userAgent), "android") {
		platform = "android"
	} else if strings.Contains(strings.ToLower(userAgent), "iphone") || strings.Contains(strings.ToLower(userAgent), "ipad") {
		platform = "ios"
	} else if strings.Contains(strings.ToLower(userAgent), "linux") {
		platform = "linux"
	} else if strings.Contains(strings.ToLower(userAgent), "macintosh") {
		platform = "mac"
	} else if strings.Contains(strings.ToLower(userAgent), "windows") {
		platform = "windows"
	}

	// Extract IP address from request context
	ipAddress := request.RequestContext.Identity.SourceIP

	_, err := dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"Shortcode": &types.AttributeValueMemberS{Value: shortcode},
			"SortKey":   &types.AttributeValueMemberS{Value: "VIEW#" + time.Now().Format(time.RFC3339)},
			"Timestamp": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
			"Action":    &types.AttributeValueMemberS{Value: "redirect"},
			"Platform":  &types.AttributeValueMemberS{Value: platform},
			"IPAddress": &types.AttributeValueMemberS{Value: ipAddress},
		},
	})
	return err
}

func CreateDynamoDBItem(shortcode string, body models.GenerateRequestBody) map[string]types.AttributeValue {
	item := map[string]types.AttributeValue{
		"Shortcode":   &types.AttributeValueMemberS{Value: shortcode},
		"SortKey":     &types.AttributeValueMemberS{Value: "META"},
		"URL":         &types.AttributeValueMemberS{Value: body.URL},
		"Description": &types.AttributeValueMemberS{Value: body.Description},
	}

	if body.AndroidURL != "" {
		item["AndroidURL"] = &types.AttributeValueMemberS{Value: body.AndroidURL}
	}
	if body.IOSURL != "" {
		item["IOSURL"] = &types.AttributeValueMemberS{Value: body.IOSURL}
	}
	if body.LinuxURL != "" {
		item["LinuxURL"] = &types.AttributeValueMemberS{Value: body.LinuxURL}
	}
	if body.MacURL != "" {
		item["MacURL"] = &types.AttributeValueMemberS{Value: body.MacURL}
	}
	if body.WindowsURL != "" {
		item["WindowsURL"] = &types.AttributeValueMemberS{Value: body.WindowsURL}
	}

	return item
}
