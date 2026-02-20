package greenlogger

// Utilities for connecting with slack

import (
	"GreenScoutBackend/constants"

	"github.com/slack-go/slack"
)

// The reference to the slack client instance
var api *slack.Client

// If the slack instance is alive
var slackAlive = false

// Initializes the instance of the slack client and stores it in memory
func InitSlackAPI(token string) bool {
	if token == "" {
		return false
	}
	api = slack.New(token)

	var validated bool
	if validated = validateToken(); validated {
		slackAlive = true
	}
	return validated
}

// Ensures the token is valid and that the client can connect to at least one workspace. If not, it will return false.
func validateToken() bool {
	res, _, err := api.ListTeams(slack.ListTeamsParameters{})

	if err != nil {
		LogError(err, "Problem listing teams the slack bot can access")
		return false
	}

	if len(res) < 1 {
		LogMessage("Slack bot cannot access any workspaces!")
		return false
	}

	return true
}

// Attempts to write a message to the channel passed in as a parameter. If there is no error, returns true.
func ValidateChannelAccess(channel string) bool {
	_, _, err := api.PostMessage(
		channel,
		slack.MsgOptionText("Spinning up server...", false),
		slack.MsgOptionAsUser(true),
	)

	return err == nil
}

// Notifies the connected slack workspace about the status of the server.
func NotifyOnline(online bool) {
	var msg string
	if online {
		msg = "ONLINE"
	} else {
		msg = "OFFLINE"
	}
	_, _, err := api.PostMessage(
		constants.CachedConfigs.SlackConfigs.Channel,
		slack.MsgOptionText("Server status: "+msg, false),
		slack.MsgOptionAsUser(true),
	)

	if err != nil {
		LogError(err, "Problem notifying server of status "+msg)
	}
}

// Sends a message to the connected slack channel
func NotifyMessage(message string) {
	_, _, postErr := api.PostMessage(
		constants.CachedConfigs.SlackConfigs.Channel,
		slack.MsgOptionText(message, false),
		slack.MsgOptionAsUser(true),
	)

	if postErr != nil {
		FatalError(postErr, "Problem writing message to slack")
	}

}

// Sends an error and its message to the connected slack channel
func NotifyError(err error, message string) {
	_, _, postErr := api.PostMessage(
		constants.CachedConfigs.SlackConfigs.Channel,
		slack.MsgOptionText("ERR: "+message, false),
		slack.MsgOptionAttachments(slack.Attachment{Text: err.Error()}),
		slack.MsgOptionAsUser(true),
	)

	if postErr != nil {
		FatalError(postErr, "Problem writing error message to slack")
	}

}

// Sets the slack Alive variable to false
func ShutdownSlack() {
	slackAlive = false
}
