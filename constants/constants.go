// Holds data in memory that is accessable project-wide.
package constants

import "path/filepath"

// The configs held in memory
var CachedConfigs GeneralConfigs

// Wether or not the event key is non-TBA
var CustomEventKey bool = false

// The filepath to the default pfp. This would be const if not for the filepath.Join()
var DefaultPfpPath = filepath.Join("pfp", "pictures", "Default_pfp.png")

// The teams from Teamlists, held in memory.
var Teams []int

// The structure of the server configurations.
type GeneralConfigs struct {
	PythonDriver       string             `yaml:"PythonDriver"`       // The driver used to run python files
	SqliteDriver       string             `yaml:"SqliteDriver"`       // The driver used to execute sqlite queries
	TBAKey             string             `yaml:"TBAKey"`             // The API Key used to connect to https://www.thebluealliance.com/apidocs/v3
	EventKey           string             `yaml:"EventKey"`           // The Blue alliance key of the event currently configured
	EventKeyName       string             `yaml:"EventKeyName"`       // The associated name of the event
	CustomEventConfigs CustomEventConfigs `yaml:"CustomEventConfigs"` // The configurations for if it is a non-TBA event
	IP                 string             `yaml:"IP"`                 // The outward-facing IPv4 address of the server
	DomainName         string             `yaml:"DomainName"`         // The domain name that matches to the server's IP
	UsingMultiScouting bool               `yaml:"UsingMultiScouting"` // If multi-scouting is enabled
	SpreadSheetID      string             `yaml:"SpreadSheetID"`      // The ID to the google sheet to be used
	PathToDatabases    string             `yaml:"PathToDatabases"`    // The filepath to the directory containing the users and authentication databases.
	SlackConfigs       SlackConfigs       `yaml:"SlackConfigs"`       // The configurations for the server's slack integration
	LogConfigs         LoggingConfigs     `yaml:"LoggingConfigs"`     // The configurations for the server's logging
}

// Configuration for slack integration
type SlackConfigs struct {
	Configured bool   `yaml:"Configured"` // If these configs have ever been generated; DO NOT EDIT THIS
	UsingSlack bool   `yaml:"UsingSlack"` // If the server will be using slack for online status notifications and error handling
	BotToken   string `yaml:"Token"`      // The slack bot token
	Channel    string `yaml:"Channel"`    // The channel which the slack bot will send error messages and status notifications to
}

type LoggingConfigs struct {
	Configured  bool `yaml:"Configured"` // If these configs have ever been generated; DO NOT EDIT THIS
	Logging     bool `yaml:"Logging"`    // If the server will be logging to GSLogs
	LoggingHttp bool `yaml:"LogHttp"`    // If the server will be logging output of the HTTP client to GSLogs
}

type CustomEventConfigs struct {
	Configured     bool `yaml:"Configured"`     // If these configs have ever been generated; DO NOT EDIT THIS
	CustomSchedule bool `yaml:"CustomSchedule"` // If there is a custom schedule.json file to be used with the custom event key
	PitScouting    bool `yaml:"PitScouting"`    // If there will be pit scouting at this event. If true, this will require an accociated file in TeamLists
}
