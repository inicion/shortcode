package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"url-shortener/models"
	"url-shortener/utils"

	"github.com/aws/aws-lambda-go/events"
)

func HandleGenerate(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var body models.GenerateRequestBody
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
			shortcode, err := utils.GenerateUniqueShortcode(ctx, body.Shortcode)
			if err != nil {
				log.Printf("Error generating shortcode: %v", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       "Failed to generate a unique shortcode",
				}, nil
			}
			// Check if the shortcode already exists
			result, err := utils.GetDynamoDBItem(ctx, shortcode)
			if err != nil {
				log.Printf("Error checking shortcode: %v", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       "Error checking shortcode",
				}, nil
			}
			if result == nil {
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

	item := utils.CreateDynamoDBItem(shortcode, body)

	if err := utils.PutDynamoDBItem(ctx, item); err != nil {
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
