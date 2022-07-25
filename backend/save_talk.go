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
	UserId string // UUID
	Talk   string
}

type talkog struct {
	UserId    string // UUID
	Timestamp int64  // Unix Time
	Talk      string
}

func createErrorResponse(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
}

func putTalkog(ctx context.Context, requestData requestData) error {
	item, err := attributevalue.MarshalMap(talkog{
		UserId:    requestData.UserId,
		Timestamp: time.Now().Unix(),
		Talk:      requestData.Talk,
	})
	if err != nil {
		return err
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("REGION")))
	if err != nil {
		return err
	}
	client := dynamodb.NewFromConfig(cfg)
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		Item:      item,
	})
	if err != nil {
		return err
	}

	return nil
}

func handle(req *events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	var requestData requestData
	err := json.Unmarshal([]byte(req.Body), &requestData)
	if err != nil {
		return createErrorResponse(err)
	}

	err = putTalkog(context.TODO(), requestData)
	if err != nil {
		return createErrorResponse(err)
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}

func main() {
	lambda.Start(handle)
}
