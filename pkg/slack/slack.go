package slack

import (
	"bytes"
	"github.com/wizedkyle/securitybot/pkg/auth"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var slackChatPostMessage string = "https://slack.com/api/chat.postMessage"

func PostToSlack(message []byte) {

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	request, err := http.NewRequest("POST", slackChatPostMessage, bytes.NewBuffer(message))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", auth.GetSlackOAuthToken())
	log.Println("Posting to Slack API: " + slackChatPostMessage)
	response, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if string(responseBody) != "ok" {
		log.Println(string(responseBody))

	}
}
