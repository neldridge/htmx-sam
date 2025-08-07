package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/a-h/templ"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	log "github.com/sirupsen/logrus"

	htmxtypes "github.com/neldridge/htmx-sam/backend/types"

	"github.com/neldridge/htmx-sam/backend/components"
)

var ddb *dynamodb.Client
var tableName = "ContactMessages"

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.Path {
	case "/home":
		return renderHTML(components.Layout("Home", components.Index()))

	case "/navbar":
		return renderHTML(components.Navbar())

	case "/contacted":
		contacted := make([]htmxtypes.ContactForm, 0)

		// Fetch contacts from DynamoDB
		resp, err := ddb.Scan(ctx, &dynamodb.ScanInput{
			TableName: &tableName,
		})
		if err != nil {
			return errorResponse(http.StatusInternalServerError, "failed to fetch contacts")
		}

		for _, item := range resp.Items {
			contact := htmxtypes.ContactForm{
				Name:    item["Name"].(*types.AttributeValueMemberS).Value,
				Email:   item["Email"].(*types.AttributeValueMemberS).Value,
				Message: item["Message"].(*types.AttributeValueMemberS).Value,
			}
			contacted = append(contacted, contact)
		}
		if len(contacted) == 0 {
			return renderHTML(components.Layout("Contacted", components.Contacted([]htmxtypes.ContactForm{})))
		}
		return renderHTML(components.Layout("Contacted", components.Contacted(contacted)))

	case "/contact":
		if req.HTTPMethod == "POST" {
			var form htmxtypes.ContactForm

			parsedForm, err := url.ParseQuery(req.Body)
			if err != nil {
				return errorResponse(http.StatusBadRequest, "invalid form data")
			}
			form.Name = parsedForm.Get("name")
			form.Email = parsedForm.Get("email")
			form.Message = parsedForm.Get("message")

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
				log.Printf("Failed to put item: %v", err)
				return errorResponse(http.StatusInternalServerError, "failed to save message")
			}
			return successResponse("Message received!")
		}
		return renderHTML(components.Layout("Contact us", components.Contact()))

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
