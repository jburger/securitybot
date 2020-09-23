package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/wizedkyle/securitybot/pkg/slackstructs"
	"log"
	"time"
)

type AccessKeys struct {
	AccessKeyId string                 `json:"AccessKeyId"`
	CreateDate  *time.Time             `json:"CreateDate"`
	Status      string                 `json:"Status"`
	UserName    string                 `json:"UserName"`
	LastUsed    AccessKeysLastUsedInfo `json:"LastUsed"`
}

type AccessKeysLastUsedInfo struct {
	LastUsedDate *time.Time `json:"LastUsedDate"`
	Region       string     `json:"Region"`
	ServiceName  string     `json:"ServiceName"`
}

type AccessKeysToDelete struct {
	UserName    string                 `json:"UserName"`
	AccessKeyId string                 `json:"AccessKeyId"`
	LastUsed    AccessKeysLastUsedInfo `json:"LastUsed"`
}

type AccessKeysToRotate struct {
	UserName    string     `json:"UserName"`
	AccessKeyId string     `json:"AccessKeyId"`
	CreateDate  *time.Time `json:"CreateDate"`
}

var accessKeysInfo []*AccessKeys
var accessKeysToDelete []*AccessKeysToDelete
var accessKeysToRotate []*AccessKeysToRotate

// Retrieves AWS access keys associated to each IAM user and access key specific information.
func getAWSKeys() {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatalln(err)
	}

	svc := iam.New(sess)

	log.Println("Getting all IAM users")
	userInput := &iam.ListUsersInput{}
	userResult, err := svc.ListUsers(userInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeServiceFailureException:
				log.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				log.Println(err.Error())
			}
		}
		return
	}

	for i, user := range userResult.Users {

		log.Println("Getting access keys for IAM user: " + *user.UserName)
		accessKeyInput := &iam.ListAccessKeysInput{
			UserName: aws.String(*user.UserName),
		}
		accessKeyResult, err := svc.ListAccessKeys(accessKeyInput)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case iam.ErrCodeNoSuchEntityException:
					log.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
				case iam.ErrCodeServiceFailureException:
					log.Println(iam.ErrCodeServiceFailureException, aerr.Error())
				default:
					log.Println(err.Error())
				}
			}
			return
		}

		for i, accessKey := range accessKeyResult.AccessKeyMetadata {

			log.Println("Getting access key last used information for Access Key ID: " + *accessKey.AccessKeyId)
			accessKeyLastUsedInput := &iam.GetAccessKeyLastUsedInput{
				AccessKeyId: accessKey.AccessKeyId,
			}
			accessKeyLastUsedResult, err := svc.GetAccessKeyLastUsed(accessKeyLastUsedInput)
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case iam.ErrCodeNoSuchEntityException:
						log.Fatalln(iam.ErrCodeNoSuchEntityException, aerr.Error())
					case iam.ErrCodeServiceFailureException:
						log.Fatalln(iam.ErrCodeServiceFailureException, aerr.Error())
					default:
						log.Fatalln(err.Error())
					}
				}
				return
			}

			userAccessKeys := new(AccessKeys)
			userAccessKeys.AccessKeyId = *accessKey.AccessKeyId
			userAccessKeys.CreateDate = accessKey.CreateDate
			userAccessKeys.Status = *accessKey.Status
			userAccessKeys.UserName = *accessKey.UserName
			userAccessKeys.LastUsed.Region = *accessKeyLastUsedResult.AccessKeyLastUsed.Region
			userAccessKeys.LastUsed.ServiceName = *accessKeyLastUsedResult.AccessKeyLastUsed.ServiceName
			userAccessKeys.LastUsed.LastUsedDate = accessKeyLastUsedResult.AccessKeyLastUsed.LastUsedDate
			accessKeysInfo = append(accessKeysInfo, userAccessKeys)
			i++
		}
		i++
	}
}

// Determines which AWS access keys need to be rotated or deleted and posts the outcome to Slack
func analyseAccessKeys() {
	now := time.Now()
	past90Days := now.AddDate(0, 0, -90)
	past365Days := now.AddDate(-1, 0, 0)
	for i, iamUser := range accessKeysInfo {
		if iamUser.Status == "Active" && iamUser.LastUsed.Region != "N/A" && iamUser.LastUsed.ServiceName != "N/A" {
			if iamUser.LastUsed.LastUsedDate.Before(past90Days) {
				toDelete := new(AccessKeysToDelete)
				toDelete.UserName = iamUser.UserName
				toDelete.AccessKeyId = iamUser.AccessKeyId
				toDelete.LastUsed.LastUsedDate = iamUser.LastUsed.LastUsedDate
				toDelete.LastUsed.ServiceName = iamUser.LastUsed.ServiceName
				toDelete.LastUsed.Region = iamUser.LastUsed.Region
				accessKeysToDelete = append(accessKeysToDelete, toDelete)
			}

			if iamUser.CreateDate.Before(past365Days) {
				toRotate := new(AccessKeysToRotate)
				toRotate.UserName = iamUser.UserName
				toRotate.AccessKeyId = iamUser.AccessKeyId
				toRotate.CreateDate = iamUser.CreateDate
				accessKeysToRotate = append(accessKeysToRotate, toRotate)
			}
		}
		i++
	}

	for i, accessKey := range accessKeysToRotate {
		var slackMessage slackstructs.SlackAWSCredCheckerStyle
		var slackMessageBlock slackstructs.SlackAWSCredCheckerBlock

		slackMessageBlock.HeaderType = "header"
		slackMessageBlock.HeaderText.Type = "plain_text"
		slackMessageBlock.HeaderText.Text = ":exclamation: You have AWS Access Keys that need to be rotated :exclamation:"
		slackMessageBlock.HeaderText.Emoji = true

		slackMessageBlock.SectionType = "section"
		slackMessageBlock.SectionText.Type = "mrkdwn"
		slackMessageBlock.SectionText.Text = "The Access Key *" + accessKey.AccessKeyId + "* is older than 1 year. What would you like to do?"

		slackMessageBlock.ActionType = "actions"
		slackMessageBlock.ActionElements.Type = "button"
		slackMessageBlock.ActionElements.ElementText.Type = "plain_text"
		slackMessageBlock.ActionElements.ElementText.Text = "Rotate Access Key"
		slackMessageBlock.ActionElements.ElementText.Emoji = true
		slackMessageBlock.ActionElements.Value = "{username:" + accessKey.UserName + ",accesskey:" + accessKey.AccessKeyId + ",action:rotate}"
		slackMessageBlock.ActionElements.Type = "button"
		slackMessageBlock.ActionElements.ElementText.Type = "plain_text"
		slackMessageBlock.ActionElements.ElementText.Text = "Delete Access Key"
		slackMessageBlock.ActionElements.ElementText.Emoji = true
		slackMessageBlock.ActionElements.Value = "{username:" + accessKey.UserName + ",accesskey:" + accessKey.AccessKeyId + ",action:delete}"
		slackMessageBlock.ActionElements.Type = "button"
		slackMessageBlock.ActionElements.ElementText.Type = "plain_text"
		slackMessageBlock.ActionElements.ElementText.Text = "Ignore Me"
		slackMessageBlock.ActionElements.ElementText.Emoji = true
		slackMessageBlock.ActionElements.Value = "{action:nothing}"

		slackMessage.Blocks = append(slackMessage.Blocks, slackMessageBlock)

		i++
	}
	// Create a DM with each user who's keys need to be rotated
	// send the actions to rotate or ignore
	// send the actions to delete or ignore
}

func LambdaHandler() {
	getAWSKeys()
	analyseAccessKeys()
}

func main() {
	lambda.Start(LambdaHandler)
}
