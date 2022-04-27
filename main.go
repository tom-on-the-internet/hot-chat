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
	"github.com/google/uuid"
)

type chatLog struct {
	messages []message
	mutex    sync.Mutex
}

type message struct {
	id     string
	author string
	body   string
}

func (c *chatLog) addMessage(author, body string) {
	c.mutex.Lock()
	uuid := uuid.NewString()
	msg := message{id: uuid, author: author, body: body}
	c.messages = append(c.messages, msg)
	c.mutex.Unlock()
}

func (c *chatLog) getMessages() []message {
	c.mutex.Lock()
	messages := c.messages
	c.mutex.Unlock()

	return messages
}

var msgLog chatLog

func main() {
	mux := chi.NewMux()
	mux.Get("/", showHandler)
	mux.Post("/", postHandler)

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

	head := `
		<head>
		  <title>Hot Chat</title>
     	  <meta charset="UTF-8">
		<link rel="icon" href="data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 100 100%22><text y=%22.9em%22 font-size=%2290%22>🔥</text></svg>">

		</head>

		<style>
		body {
          background-color: #ff0000;
          background-image: linear-gradient(315deg, #ff0000 0%, #ffed00 74%);
		}
		main {
	      display: flex;
	    }
		main>div {
		  border: solid 1px black;
		  padding: 10px;
	    }
	    #chat-log{
          flex:1;
      	}
        .author{
          text-decoration: underline;
        }
        .author>span{
          font-weight: bold;
        }

		</style>
    `
	page := `

		<h1>Hot Chat</h1>
	    <main>
	    <div>
	      <form action="/" method="post">
	        <label for="author">
              Your Name:
	        </label>
	        </br>
	        <input required id="author" name="author" type="text"/>
	        </br>
	        </br>
	        <label for="message">
              Message:
	        </label>
	        </br>
	        <input required id="message" name="message" type="text"/>
	        </br>
	        </br>
	        <button type="submit">Send</button>
	      </form>
	    </div>
	    <div id="chat-log">
	      <h2>Chat Log</h2>
	      %s
	    </div>
    	</main>
    	<script>
          const $author = document.getElementById('author')
          const author = localStorage.getItem('author');
          $author.value = author;
          $author.addEventListener('change', function(event){
            localStorage.setItem('author', $author.value);
          });
    	</script>
`

	messages := ""

	for _, message := range msgLog.getMessages() {
		element := `
        <div id="%s">
        <div class="author"><span>%s</span> wrote:</div>
        <div class="message-body">"%s"</div>
        <hr>
        </div>
`
		element = fmt.Sprintf(element, message.id, message.author, message.body)
		messages += element
	}

	page = head + fmt.Sprintf(page, messages)

	_, err := w.Write([]byte(page))
	if err != nil {
		log.Println(err)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	author := r.FormValue("author")
	message := r.FormValue("message")

	msgLog.addMessage(author, message)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
