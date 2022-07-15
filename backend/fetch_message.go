package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

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

type record struct {
	UserId    string // UUID
	Message   string
	Timestamp int64 // Unix Time
}

func createErrorResponse(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
}

func handle(ctx context.Context, req *events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	var requestData requestData
	err := json.Unmarshal([]byte(req.Body), &requestData)
	if err != nil {
		return createErrorResponse(err)
	}

	conf, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("REGION")))
	if err != nil {
		return createErrorResponse(err)
	}
	// TODO: use DynamoDB Stream
	clientoo := dynamodb.NewFromConfig(conf)
	expr, err := expression.NewBuilder().WithFilter(
		expression.Contains(expression.Name("UserId"), requestData.UserId),
	).Build()
	if err != nil {
		return createErrorResponse(err)
	}
	scan, err := clientoo.Scan(ctx, &dynamodb.ScanInput{
		TableName:                 aws.String(os.Getenv("DYNAMODB_TABLE")),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return createErrorResponse(err)
	}
	var records []record
	err = attributevalue.UnmarshalListOfMaps(scan.Items, &records)
	if err != nil {
		return createErrorResponse(err)
	}

	// *************************************************

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
		return createErrorResponse(err)
	}

	client := apigatewaymanagementapi.NewFromConfig(cfg)
	for _, record := range records {
		data, err := json.Marshal(record.Message)
		if err != nil {
			return createErrorResponse(err)
		}
		_, err := client.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: &req.RequestContext.ConnectionID,
			Data:         data,
		})
		if err != nil {
			return createErrorResponse(err)
		}
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}

func main() {
	lambda.Start(handle)
}
