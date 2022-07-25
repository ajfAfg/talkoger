package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/ajfAfg/talkoger/backend/talkog"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type requestData struct {
	UserId string // UUID
	Talk   string
}

type connection struct {
	ConnectionId string
	UserId       string // UUID
}

func createErrorResponse(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
}

func putTalkog(ctx context.Context, talkog talkog.Talkog) error {
	item, err := attributevalue.MarshalMap(talkog)
	if err != nil {
		return err
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("REGION")))
	if err != nil {
		return err
	}
	client := dynamodb.NewFromConfig(cfg)
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("TALKOG_TABLE")),
		Item:      item,
	})
	if err != nil {
		return err
	}

	return nil
}

func getConnectionsByUserId(ctx context.Context, userId string) ([]connection, error) {
	var connections []connection

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("REGION")))
	if err != nil {
		return connections, err
	}
	client := dynamodb.NewFromConfig(cfg)

	expr, err := expression.NewBuilder().WithFilter(
		expression.Equal(expression.Name("UserId"), expression.Value(userId)),
	).Build()
	if err != nil {
		return connections, err
	}
	scan, err := client.Scan(ctx, &dynamodb.ScanInput{
		TableName:                 aws.String(os.Getenv("CONNECTION_TABLE")),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return connections, err
	}

	err = attributevalue.UnmarshalListOfMaps(scan.Items, &connections)
	if err != nil {
		return connections, err
	}

	return connections, nil
}

func sendTalkog(ctx context.Context, talkog talkog.Talkog, userId string, req *events.APIGatewayWebsocketProxyRequest) error {
	connections, err := getConnectionsByUserId(ctx, userId)
	if err != nil {
		return err
	}

	endpoint := url.URL{
		Path:   req.RequestContext.Stage,
		Host:   req.RequestContext.DomainName,
		Scheme: "https",
	}
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithEndpointResolver(
			aws.EndpointResolverFunc(
				func(service, region string) (aws.Endpoint, error) {
					return aws.Endpoint{
						SigningRegion: os.Getenv("REGION"),
						URL:           endpoint.String(),
					}, nil
				})),
	)
	if err != nil {
		return err
	}

	client := apigatewaymanagementapi.NewFromConfig(cfg)
	data, err := json.Marshal(talkog)
	if err != nil {
		return err
	}
	for _, conn := range connections {
		_, err = client.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: &conn.ConnectionId,
			Data:         data,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func handle(ctx context.Context, req *events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	var requestData requestData
	err := json.Unmarshal([]byte(req.Body), &requestData)
	if err != nil {
		return createErrorResponse(err)
	}

	talkog := talkog.New(requestData.UserId, requestData.Talk)

	err = putTalkog(ctx, talkog)
	if err != nil {
		return createErrorResponse(err)
	}

	err = sendTalkog(ctx, talkog, requestData.UserId, req)
	if err != nil {
		return createErrorResponse(err)
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}

func main() {
	lambda.Start(handle)
}
