package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type requestData struct {
	UserId  string // UUID
	Message string
}

type saveData struct {
	UserId    string // UUID
	Message   string
	Timestamp int64 // Unix Time
}

func createErrorResponse(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
}

func handle(req *events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	var requestData requestData
	err := json.Unmarshal([]byte(req.Body), &requestData)
	if err != nil {
		return createErrorResponse(err)
	}

	item, err := attributevalue.MarshalMap(saveData{
		UserId:    requestData.UserId,
		Message:   requestData.Message,
		Timestamp: time.Now().Unix(),
	})
	if err != nil {
		return createErrorResponse(err)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("REGION")))
	if err != nil {
		return createErrorResponse(err)
	}
	svc := dynamodb.NewFromConfig(cfg)
	_, err = svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		Item:      item,
	})
	if err != nil {
		return createErrorResponse(err)
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}

func main() {
	lambda.Start(handle)
}
