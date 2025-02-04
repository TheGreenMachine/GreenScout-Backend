package lib

// Utility for parsing and processing match JSON

import (
	"GreenScoutBackend/constants"
	greenlogger "GreenScoutBackend/greenLogger"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Data from one scouter from one match
type TeamData struct {
	TeamNumber    uint64            `json:"Team"`             // The team number
	Match         MatchInfo         `json:"Match"`            // The match number
	Scouter       string            `json:"Scouter"`          // The scouter who recorded this data
	DriverStation DriverStationData `json:"Driver Station"`   // The driver station
	Cycles        []Cycle           `json:"Cycles"`           // The cycle data
	Pickups       PickupLocations   `json:"Pickup Locations"` // The recorded speaker locations
	Auto          AutoData          `json:"Auto"`             // The autonomous data
	Climb         ClimbingData      `json:"Climbing"`         // The recorded climbing data
	Trap          TrapData          `json:"Trap"`             // The recorded trap data
	Endgame       EndgameData       `json:"Endgame"`          // The recorded endgame data
	Misc          MiscData          `json:"Misc"`             // Miscellaneous data
	Penalties     []string          `json:"Penalties"`        // Recorded penalties
	Rescouting    bool              `json:"Rescouting"`       // If this match is rescouting (Will override all previous data of this match with this driverstation)
	Notes         string            `json:"Notes"`            // Notes from the scouter
}

// Basic info about the match
type MatchInfo struct {
	Number   uint `json:"Number"`   // The match number
	IsReplay bool `json:"isReplay"` // If it is a replay
}

// Basic info about the driver station
type DriverStationData struct {
	IsBlue bool `json:"Is Blue"` // If it is blue
	Number int  `json:"Number"`  // The driverstation number (1-3)
}

// One cycle
type Cycle struct {
	Time    float64 `json:"Time"`    // The time taken
	Type    string  `json:"Type"`    // The type of cycle
	Success bool    `json:"Success"` // If it was successful
}

// Data from the autonomous period
type AutoData struct {
	Can    bool `json:"Can"`    // If the robot has/can do autonomous
	Scores int  `json:"Scores"` // The scores in auto
	Misses int  `json:"Misses"` // The misses in auto
	Ejects int  `json:"Ejects"` // The ejects/shuttles in auto
}

// Data about the climb of a robot
type ClimbingData struct {
	Succeeded bool    `json:"Succeeded"` // If it successfully climbed
	Time      float64 `json:"Time"`      // How long it took to climb
}

// Data about a robot's performance at the trap
type TrapData struct {
	Attempts int `json:"Attempts"` // The attempts it took
	Score    int `json:"Score"`    // The number of notes it scored
}

// Data about a robot's performance during parking, currently just for parking atm
type EndgameData struct {
	parkStatus int `json:"Parking Status"` // What the robot at the end of the game (i.e. did it park, did it climb, etc)

}

// Where a robot could pick up from
type PickupLocations struct {
	CoralGround bool `json:"Coral Ground"` // If it could pick up from the ground
	CoralSource bool `json:"Coral Source"` // If it could pick up from the source
	AlgaeGround bool `json:"Algae Ground"` // If it could pick up from the ground
	AlgaeSource bool `json:"Algae Source"` // If it could pick up from the source

}

// Miscellaneous robot data
type MiscData struct {
	DC        bool `json:"Lost Communication"` // If the robot DC'd
	LostTrack bool `json:"User Lost Track"`    // If the scouter lost track
	Disabled  bool `json:"Disabled"`           // If the robot was disabled
}

// Parses through the file at the passed in location, returning a compiled TeamData object and wether or not there were errors.
// Params: The filepath, if it has already been written (for multi-scouting)
func Parse(file string, hasBeenWritten bool) (TeamData, bool) {

	var path string
	if hasBeenWritten {
		path = filepath.Join(constants.JsonWrittenDirectory, file)
	} else {
		path = filepath.Join(constants.JsonInDirectory, file)
	}

	// Open file
	jsonFile, fileErr := os.Open(path)

	// Handle any error opening the file
	if fileErr != nil {
		greenlogger.LogErrorf(fileErr, "Error opening JSON file %v", path)
		return TeamData{}, true
	}

	// defer file closing
	defer jsonFile.Close()

	var teamData TeamData

	dataAsByte, readErr := io.ReadAll(jsonFile)

	if readErr != nil {
		greenlogger.LogErrorf(readErr, "Error reading JSON file %v", path)
		return TeamData{}, true
	}

	//Deocding
	err := json.Unmarshal(dataAsByte, &teamData)

	//Deal with unmarshalling errors
	if err != nil {
		greenlogger.LogErrorf(err, "Error unmarshalling JSON data %v", string(dataAsByte))
		return TeamData{}, true
	}

	return teamData, false
}

// Identifying information on one driverstation on one match.
// Used for the GETSCOUTER() method in the spreadsheet.
type MatchInfoRequest struct {
	Match         int  `json:"Match"`         // The match number
	IsBlue        bool `json:"isBlue"`        // If the driverstation is blue
	DriverStation int  `json:"DriverStation"` // The driverstation number
}

// Matches the parameters of the passed in MatchInfoRequest and returns all scouters who scouted that match.
func GetNameFromWritten(match MatchInfoRequest) string {
	var names []string

	filePattern := fmt.Sprintf("%s_%v_%s", GetCurrentEvent(), match.Match, GetDSString(match.IsBlue, uint(match.DriverStation)))

	written, err := os.ReadDir(constants.JsonWrittenDirectory)

	if err != nil {
		greenlogger.LogErrorf(err, "Error searching %v", constants.JsonWrittenDirectory)
		return "Err in searching!"
	}

	for _, file := range written {

		splitByUnder := strings.Split(file.Name(), "_")

		fmt.Printf("%v", splitByUnder)

		if len(splitByUnder) > 3 && filePattern == strings.Join(splitByUnder[:3], "_") {

			// Open file
			outFilePath := filepath.Join(constants.JsonWrittenDirectory, file.Name())
			jsonFile, fileErr := os.Open(outFilePath)

			// Handle any error opening the file
			if fileErr != nil {
				greenlogger.LogErrorf(fileErr, "Error opening JSON file %v", outFilePath)
			}

			// defer file closing
			defer jsonFile.Close()

			var teamData TeamData

			dataAsByte, readErr := io.ReadAll(jsonFile)

			if readErr != nil {
				greenlogger.LogErrorf(readErr, outFilePath)
			}

			//Deocding
			err := json.Unmarshal(dataAsByte, &teamData)

			//Deal with unmarshalling errors
			if err != nil {
				greenlogger.LogErrorf(err, "Error unmarshalling JSON data %v", string(dataAsByte))
			}

			if teamData.Scouter != "" {
				names = append(names, teamData.Scouter)

			}
		}
	}

	if len(names) == 0 {
		return "No scouters found!"
	}

	return strings.Join(names, ", ")
}

//!!! PIT SCOUTING IS NOT YET IMPLEMENTED ON THE FRONTEND !!!//

// Data from pit scouting
type PitScoutingData struct {
	TeamNumber    int    `json:"Team"`    // The team number
	PitIdentifier string `json:"Pit"`     // The pit identifier, as seen on the pit map
	Scouter       string `json:"Scouter"` // The person who did the pit scouting

	Drivetrain string       `json:"Drivetrain"` // The type of drivetrain the robot has
	Distance   DistanceData `json:"Distance"`   // The information on the distance shooting it can do

	AutoScores     int  `json:"Auto Scores"`  // The average scores this robot gets in auto
	MiddleControls int  `json:"Middle Notes"` // The average number of notes from the middle this robot gets in auto
	NoteDetection  bool `json:"Detection"`    // If it has note detection

	Cycles           int             `json:"Cycles"`       // The average number of cycles it gets in teleop
	DriverExperience int             `json:"Experience"`   // The years of experience the driver has
	BotType          string          `json:"Bot Type"`     // The type of robot it is (offense, defense, amp, speaker, etc)
	HumanPlayer      HumanPlayerData `json:"Human Player"` // Data regarding this team's human player

	EndgameBehavior string  `json:"Endgame Behavior"` // What does this robot do in endgame? (Park, climb, trap, etc)
	ClimbTime       float64 `json:"Climb Time"`       // How long does it take for this robot to climb?

	Notes string `json:"Notes"` // Other notes
}

// Pit scouting data regarding distance shooting
type DistanceData struct {
	Can      bool    `json:"Can"`      // Do they say they can distance shoot?
	Distance float64 `json:"Distance"` // How many feet away they can shoot from
}

// Pit scouting data regarding the human player
type HumanPlayerData struct {
	Position      int `json:"Position"`       // What position the human player prefers (source, amp, etc)
	StageAccuracy int `json:"Stage Accuracy"` // How accurate the human player is at throwing the note onto the stage (sorry elena)
}

// Parses through the file at the passed in location, returning a compiled PitScoutingData object and wether or not there were errors.
func ParsePitScout(file string) (PitScoutingData, bool) {

	path := filepath.Join(constants.JsonInDirectory, file)

	// Open file
	jsonFile, fileErr := os.Open(path)

	// Handle any error opening the file
	if fileErr != nil {
		greenlogger.LogErrorf(fileErr, "Error opening JSON file %v", path)
		return PitScoutingData{}, true
	}

	// defer file closing
	defer jsonFile.Close()

	var pitData PitScoutingData

	dataAsByte, readErr := io.ReadAll(jsonFile)

	if readErr != nil {
		greenlogger.LogErrorf(readErr, "Error reading JSON file %v", path)
		return PitScoutingData{}, true
	}

	//Deocding
	err := json.Unmarshal(dataAsByte, &pitData)
	//Deal with unmarshalling errors
	if err != nil {
		greenlogger.LogErrorf(err, "Error unmarshalling JSON data %v", string(dataAsByte))
		return PitScoutingData{}, true
	}

	return pitData, false
}
