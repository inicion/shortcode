package main

import (
	"log"

	"url-shortener/handlers"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	log.Println("Starting Lambda function")
	lambda.Start(handlers.HandleRequest)
}
