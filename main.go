package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/go-chi/chi/v5"
)

type chatLog struct {
	messages [][2]string
	mutex    sync.Mutex
}

func (c *chatLog) addMessage(author, message string) {
	c.mutex.Lock()
	msg := [2]string{author, message}
	c.messages = append(c.messages, msg)
	c.mutex.Unlock()
}

func (c *chatLog) getMessages() [][2]string {
	c.mutex.Lock()
	messages := c.messages
	c.mutex.Unlock()

	return messages
}

var msgLog chatLog

func main() {
	mux := chi.NewMux()
	mux.Get("/", showHandler)
	mux.Post("/truth", postHandler)

	log.Println("🔥 Hot Chat")

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") == "" {
		// local
		log.Fatal(http.ListenAndServe(":3000", mux))
	} else {
		// lambda
		lambda.Start(httpadapter.New(mux).ProxyWithContext)
	}
}

func showHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	msgLog.addMessage("Keith", "Hi, I'm Keith.")

	messages := ""

	for _, message := range msgLog.getMessages() {
		element := `
        <div>
          <div class="author">%s</div>
          <div class="message">%s</div>
        </div>
`
		element = fmt.Sprintf(element, message[0], message[1])
		messages += element
	}

	page := `
		<head>
		  <title>Hot Chat</title>
     	  <meta charset="UTF-8">
		<link rel="icon" href="data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 100 100%22><text y=%22.9em%22 font-size=%2290%22>🔥</text></svg>">

		</head>

		<style>
		body {

		}
		</style>

		<h1>Hot Chat</h1>
`

	page += messages

	_, err := w.Write([]byte(page))
	if err != nil {
		log.Println(err)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
}
