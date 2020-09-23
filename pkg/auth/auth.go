package auth

import (
	"os"
)

func GetSlackOAuthToken() string {
	oauthToken := os.Getenv("SlackOAuthToken")
	return oauthToken
}

func ValidateSlackRequest() {

}
