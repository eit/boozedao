package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	_ "github.com/joho/godotenv/autoload"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// fmt.Println("req:", req.Headers["x-line-signature"])
	// fmt.Println("ctx:", ctx)
	// fmt.Println("This message will show up in the CLI console.")
	// fmt.Println(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_ACCESS_TOKEN"))
	bot, err := linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_ACCESS_TOKEN"))
	if err != nil {
		fmt.Println(err)
	}

	r := proxyRequest2httpRequest(&req)
	event_set, err := bot.ParseRequest(r)
	if err != nil {
		if errors.Is(err, linebot.ErrInvalidSignature) {
			return &events.APIGatewayProxyResponse{
				StatusCode:      400,
				Headers:         map[string]string{"Content-Type": "text/plain"},
				IsBase64Encoded: false,
			}, nil
		} else {
			return &events.APIGatewayProxyResponse{
				StatusCode:      500,
				Headers:         map[string]string{"Content-Type": "text/plain"},
				IsBase64Encoded: false,
			}, nil
		}
	}

	for _, event := range event_set {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			// Handle only on text message
			case *linebot.TextMessage:

			// Handle only on Sticker message
			case *linebot.StickerMessage:
				var kw string
				for _, k := range message.Keywords {
					kw = kw + "," + k
				}

				outStickerResult := fmt.Sprintf("收到貼圖訊息: %s, pkg: %s kw: %s  text: %s", message.StickerID, message.PackageID, kw, message.Text)
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(outStickerResult)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}

	return &events.APIGatewayProxyResponse{
		StatusCode:      200,
		Headers:         map[string]string{"Content-Type": "text/plain"},
		Body:            "Hello, world!",
		IsBase64Encoded: false,
	}, nil
}

func proxyRequest2httpRequest(request *events.APIGatewayProxyRequest) *http.Request {
	httpRequest, err := http.NewRequest(
		strings.ToUpper(request.HTTPMethod),
		request.Path,
		bytes.NewReader([]byte(request.Body)),
	)
	if err != nil {
		fmt.Printf("Convert To Request Failed")
	}

	if request.MultiValueHeaders != nil {
		for k, values := range request.MultiValueHeaders {
			for _, value := range values {
				httpRequest.Header.Add(k, value)
			}
		}
	} else {
		for h := range request.Headers {
			httpRequest.Header.Add(h, request.Headers[h])
		}
	}

	httpRequest.RequestURI = httpRequest.URL.RequestURI()

	return httpRequest
}

func main() {
	lambda.Start(handler)
}
