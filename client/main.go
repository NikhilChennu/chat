package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"net/http"

	"github.com/gorilla/websocket"
)

var (
	port   = flag.String("port", "9000", "port used for ws connection")
	userId = flag.String("id", "Guest", "id used for ws connection")
	mobile = flag.String("mobile", "9949071409", "mobile number")
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func main() {
	flag.Parse()
	requestBody, err := json.Marshal(map[string]string{"userId": *userId, "mobile": *mobile})

	if err != nil {
		fmt.Errorf("Error forming Json")
	}

	resp, err := http.Post("http://localhost:9000/create", "application/json", bytes.NewBuffer(requestBody))

	if err != nil {
		fmt.Errorf(err.Error())
	}

	req, err := http.NewRequest("POST", "http://localhost:9000/login", bytes.NewBuffer(requestBody))
	httpClient := http.Client{
		Transport: http.DefaultTransport,
	}

	if err != nil {
		fmt.Errorf(err.Error())
	}
	resp, err = httpClient.Do(req)

	if reqHeadersBytes, err := json.Marshal(resp.Header); err != nil {
		log.Println("Could not Marshal Req Headers")
	} else {
		fmt.Println(string(reqHeadersBytes))
	}

	// connect
	ws, _, err := connect(resp.Header)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// receive
	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				break
			}
			message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
			fmt.Println("Message: ", string(message))
		}
	}()

	// send
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}

		err := ws.WriteMessage(websocket.TextMessage, []byte(text))
		if err != nil {
			return
		}
	}
}

// connect connects to the local chat server at port <port>
func connect(header http.Header) (*websocket.Conn, *http.Response, error) {
	return websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:%s/chat", *port), header)
}
