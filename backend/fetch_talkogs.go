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
}

type connection struct {
	ConnectionId string
	UserId       string // UUID
}

func createErrorResponse(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
}

func putConnection(ctx context.Context, connectionId string, userId string) error {
	item, err := attributevalue.MarshalMap(connection{
		ConnectionId: connectionId,
		UserId:       userId,
	})
	if err != nil {
		return err
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("REGION")))
	if err != nil {
		return err
	}
	client := dynamodb.NewFromConfig(cfg)
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("CONNECTION_TABLE")),
		Item:      item,
	})
	if err != nil {
		return err
	}

	return nil
}

func getTalkogsByUserId(ctx context.Context, userId string) ([]talkog.Talkog, error) {
	var records []talkog.Talkog

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("REGION")))
	if err != nil {
		return records, err
	}
	client := dynamodb.NewFromConfig(cfg)

	expr, err := expression.NewBuilder().WithFilter(
		expression.Equal(expression.Name("UserId"), expression.Value(userId)),
	).Build()
	if err != nil {
		return records, err
	}
	scan, err := client.Scan(ctx, &dynamodb.ScanInput{
		TableName:                 aws.String(os.Getenv("TALKOG_TABLE")),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return records, err
	}

	err = attributevalue.UnmarshalListOfMaps(scan.Items, &records)
	if err != nil {
		return records, err
	}

	return records, nil
}

func sendTalkogs(ctx context.Context, talkogs []talkog.Talkog, req *events.APIGatewayWebsocketProxyRequest) error {
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
	for _, record := range talkogs {
		data, err := json.Marshal(record)
		if err != nil {
			return err
		}
		_, err = client.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: &req.RequestContext.ConnectionID,
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

	err = putConnection(ctx, req.RequestContext.ConnectionID, requestData.UserId)
	if err != nil {
		return createErrorResponse(err)
	}

	talkogs, err := getTalkogsByUserId(ctx, requestData.UserId)
	if err != nil {
		return createErrorResponse(err)
	}

	err = sendTalkogs(ctx, talkogs, req)
	if err != nil {
		return createErrorResponse(err)
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}

func main() {
	lambda.Start(handle)
}
