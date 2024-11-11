package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"url-shortener/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func HandleRedirect(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	shortcode, exists := request.PathParameters["code"]
	if !exists || shortcode == "" {
		log.Println("Missing 'code' parameter")
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Missing 'code' parameter",
		}, nil
	}

	item, err := utils.GetDynamoDBItem(ctx, shortcode)
	if err != nil || item == nil {
		log.Printf("Shortcode not found: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "Shortcode not found",
		}, nil
	}

	longURL := item["URL"].(*types.AttributeValueMemberS).Value

	// Check the User-Agent header to determine the device type
	userAgent := request.Headers["User-Agent"]
	if strings.Contains(strings.ToLower(userAgent), "android") && item["AndroidURL"] != nil {
		longURL = item["AndroidURL"].(*types.AttributeValueMemberS).Value
	} else if (strings.Contains(strings.ToLower(userAgent), "iphone") || strings.Contains(strings.ToLower(userAgent), "ipad")) && item["IOSURL"] != nil {
		longURL = item["IOSURL"].(*types.AttributeValueMemberS).Value
	} else if strings.Contains(strings.ToLower(userAgent), "linux") && item["LinuxURL"] != nil {
		longURL = item["LinuxURL"].(*types.AttributeValueMemberS).Value
	} else if strings.Contains(strings.ToLower(userAgent), "macintosh") && item["MacURL"] != nil {
		longURL = item["MacURL"].(*types.AttributeValueMemberS).Value
	} else if strings.Contains(strings.ToLower(userAgent), "windows") && item["WindowsURL"] != nil {
		longURL = item["WindowsURL"].(*types.AttributeValueMemberS).Value
	}

	if err := utils.LogRedirect(ctx, shortcode, request, longURL); err != nil {
		log.Printf("Error logging usage: %v", err)
		// Don't return an error to the user if logging fails
		// return events.APIGatewayProxyResponse{
		// 	StatusCode: http.StatusInternalServerError,
		// 	Body:       "Error logging usage",
		// }, nil
	}

	log.Printf("Redirecting to: %s", longURL)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusFound,
		Headers: map[string]string{
			"Location": longURL,
		},
	}, nil
}
