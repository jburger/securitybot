package auth

import (
	"os"
)

func GetSlackOAuthToken() string {
	oauthToken := os.Getenv("SlackOAuthToken")
	bearerToken := "Bearer " + oauthToken
	return bearerToken
}

func ValidateSlackRequest() {

}
