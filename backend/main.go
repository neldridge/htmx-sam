package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	components "github.com/neldridge/htmx-sam/backend/components"
)

type ContactForm struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

var ddb *dynamodb.Client
var tableName = "ContactMessages"

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.Path {
	case "/":
		component := components.Layout("Home", components.Index())
		return renderHTML(component)
	case "/contact":
		if req.HTTPMethod == "POST" {
			var form ContactForm
			err := json.Unmarshal([]byte(req.Body), &form)
			if err != nil {
				return errorResponse(http.StatusBadRequest, "invalid input")
			}

			item := map[string]types.AttributeValue{
				"Email":   &types.AttributeValueMemberS{Value: form.Email},
				"Name":    &types.AttributeValueMemberS{Value: form.Name},
				"Message": &types.AttributeValueMemberS{Value: form.Message},
			}
			_, err = ddb.PutItem(ctx, &dynamodb.PutItemInput{
				TableName: &tableName,
				Item:      item,
			})
			if err != nil {
				return errorResponse(http.StatusInternalServerError, "failed to save message")
			}
			return successResponse("Message received!")
		}
		return renderHTML(components.Contact())
	default:
		return errorResponse(http.StatusNotFound, "not found")
	}
}

func renderHTML(c templ.Component) (events.APIGatewayProxyResponse, error) {
	h := http.Header{}
	h.Set("Content-Type", "text/html")
	w := &responseWriter{headers: h, body: []byte{}, statusCode: 200}
	c.Render(context.Background(), w)
	return events.APIGatewayProxyResponse{
		StatusCode: w.statusCode,
		Headers:    map[string]string{"Content-Type": "text/html"},
		Body:       string(w.body),
	}, nil
}

type responseWriter struct {
	headers    http.Header
	body       []byte
	statusCode int
}

func (rw *responseWriter) Header() http.Header { return rw.headers }
func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return len(b), nil
}
func (rw *responseWriter) WriteHeader(statusCode int) { rw.statusCode = statusCode }

func successResponse(msg string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       msg,
	}, nil
}

func errorResponse(code int, msg string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Body:       msg,
	}, nil
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	ddb = dynamodb.NewFromConfig(cfg)
	lambda.Start(handler)
}
