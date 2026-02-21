// Holds data in memory that is accessable project-wide.
package internal

import (
	"path/filepath"
)

// The constant reference to the setup yaml
var ConfigFilePath = filepath.Join("conf", "greenscout.config.yaml")

// The configs held in memory
var CachedConfigs GeneralConfigs
var JsonInDirectory string
var JsonWrittenDirectory string
var JsonMangledDirectory string
var JsonArchiveDirectory string
var JsonErroredDirectory string
var JsonDiscardedDirectory string
var JsonPitWrittenDirectory string

// The default domain to be allowed to query the backend for CORS
// (e.g. the frontend). This allows *all* domains to query the backend, and
// thus should *always* be set in the configuration file when in production.
var DefaultFrontendDomain = "*"

// Wether or not the event key is non-TBA
var CustomEventKey bool = false

// The filepath to the default pfp
var DefaultPfp = "Default_pfp.png"

var DefaultRuntimeDirectory = "run"
var DefaultPfpDirectory = "pfp"
var DefaultGalleryDirectory = "gallery"
var DefaultJsonDirectory = "json"
var DefaultDbDirectory = "GreenScout-Databases"
var DefaultLogDirectory = "logs"
var DefaultTeamsDirectory = "teams"
var DefaultCertsDirectory = "certs"

var EnableHttps = true

var RSAPubKeyPath string
var RSAPrivateKeyPath string
var SheetsTokenFile string
var DefaultPfpPath string

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
	FrontendDomain     string             `yaml:"FrontendDomain"`     // The domain hosting the GreenScout frontend (for CORS)
	UsingMultiScouting bool               `yaml:"UsingMultiScouting"` // If multi-scouting is enabled
	SpreadSheetID      string             `yaml:"SpreadSheetID"`      // The ID to the google sheet to be used
	PathToDatabases    string             `yaml:"PathToDatabases"`    // The filepath to the directory containing the users and authentication databases.
	RuntimeDirectory   string             `yaml:"RuntimeDirectory"`
	JsonDirectory      string             `yaml:"JsonDirectory"`
	TeamListsDirectory string             `yaml:"TeamListsDirectory"`
	PfpDirectory       string             `yaml:"PfpDirectory"`
	GalleryDirectory   string             `yaml:"GalleryDirectory"`
	CertsDirectory     string             `yaml:"CertsDirectory"`
	LogConfigs         LoggingConfigs     `yaml:"LoggingConfigs"` // The configurations for the server's logging
}

type LoggingConfigs struct {
	Configured  bool `yaml:"Configured"` // If these configs have ever been generated; DO NOT EDIT THIS
	Logging     bool `yaml:"Logging"`    // If the server will be logging to GSLogs
	LoggingHttp bool `yaml:"LogHttp"`    // If the server will be logging output of the HTTP client to GSLogs
}

type CustomEventConfigs struct {
	Configured     bool `yaml:"Configured"`     // If these configs have ever been generated; DO NOT EDIT THIS
	CustomSchedule bool `yaml:"CustomSchedule"` // If there is a custom json file to be used with the custom event key
	PitScouting    bool `yaml:"PitScouting"`    // If there will be pit scouting at this event. If true, this will require an accociated file in TeamLists
}
