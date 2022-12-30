package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	_ "github.com/joho/godotenv/autoload"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	fmt.Println("req:", req)
	log.Println("req log:", req)
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

	messages := []string{
		"愛你",
		"愛你啦",
		"好啦愛你啦",
		"<3",
	}

	for _, event := range event_set {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			// Handle only on text message
			case *linebot.TextMessage:
				var reply string
				if strings.Contains(message.Text, "口味") {
					reply = "西西里開心果\n水梨油菊\n大便巧克力\n撒尿牛丸\n" + "每種口味都好棒棒，搭配調酒更是爽到不行"
				} else if strings.Contains(message.Text, "介紹") {
					reply = "我們是 Gelato Bar 所以有 Gelato 也有 Wine/Cocktail and we have coffee as well"
				} else {
					rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
					reply = messages[rand.Intn(len(messages))]
				}

				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
					log.Print(err)
				}

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
