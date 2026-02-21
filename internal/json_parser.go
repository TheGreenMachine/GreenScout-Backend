package internal

// Utility for parsing and processing match JSON

import (
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
	Auto          AutoData          `json:"Auto"`             // The autonomous data           // The recorded trap data
	Endgame       EndgameData       `json:"Endgame"`          // The recorded endgame data
	Misc          MiscData          `json:"Misc"`             // Miscellaneous data
	Penalties     []string          `json:"Penalties"`        // Recorded penalties
	Rescouting    bool              `json:"Rescouting"`       // If this match is rescouting (Will override all previous data of this match with this driverstation)
	Prescouting   bool              `json:"Prescouting"`      // If this match is prescouting (removes match req for saving)
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

// Where a robot could pick up from
type PickupLocations struct {
	CoralGround bool `json:"Coral Ground"` // If it could pick up from the ground
	CoralSource bool `json:"Coral Source"` // If it could pick up from the source
	AlgaeGround bool `json:"Algae Ground"` // If it could pick up from the ground
	AlgaeSource bool `json:"Algae Source"` // If it could pick up from the source
}

// Data from the autonomous period
type AutoData struct {
	Can    bool `json:"Can"`    // If the robot has/can do autonomous
	Scores int  `json:"Scores"` // The scores in auto
	Misses int  `json:"Misses"` // The misses in auto
	Ejects int  `json:"Ejects"` // The ejects/shuttles in auto
}

// Data about a robot's performance during parking, currently just for parking atm
type EndgameData struct {
	ParkStatus int     `json:"Parking Status"` // What the robot at the end of the game (i.e. did it park, did it climb, etc)
	Time       float64 `json:"Time"`           // How long it took to climb
}

// Miscellaneous robot data
type MiscData struct {
	DC        bool `json:"Lost Communication or Disabled"` // If the robot DC'd
	LostTrack bool `json:"User Lost Track"`                // If the scouter lost track
}

// Parses through the file at the passed in location, returning a compiled TeamData object and wether or not there were errors.
// Params: The filepath, if it has already been written (for multi-scouting)
func Parse(file string, hasBeenWritten bool) (TeamData, bool) {

	var path string
	if hasBeenWritten {
		path = filepath.Join(JsonWrittenDirectory, file)
	} else {
		path = filepath.Join(JsonInDirectory, file)
	}

	// Open file
	jsonFile, fileErr := os.Open(path)

	// Handle any error opening the file
	if fileErr != nil {
		LogErrorf(fileErr, "Error opening JSON file %v", path)
		return TeamData{}, true
	}

	// defer file closing
	defer jsonFile.Close()

	var teamData TeamData

	dataAsByte, readErr := io.ReadAll(jsonFile)

	if readErr != nil {
		LogErrorf(readErr, "Error reading JSON file %v", path)
		return TeamData{}, true
	}

	//Deocoding
	err := json.Unmarshal(dataAsByte, &teamData)

	//Deal with unmarshalling errors
	if err != nil {
		LogErrorf(err, "Error unmarshalling JSON data %v", string(dataAsByte))
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

	written, err := os.ReadDir(JsonWrittenDirectory)

	if err != nil {
		LogErrorf(err, "Error searching %v", JsonWrittenDirectory)
		return "Err in searching!"
	}

	for _, file := range written {

		splitByUnder := strings.Split(file.Name(), "_")

		fmt.Printf("%v", splitByUnder)

		if len(splitByUnder) > 3 && filePattern == strings.Join(splitByUnder[:3], "_") {

			// Open file
			outFilePath := filepath.Join(JsonWrittenDirectory, file.Name())
			jsonFile, fileErr := os.Open(outFilePath)

			// Handle any error opening the file
			if fileErr != nil {
				LogErrorf(fileErr, "Error opening JSON file %v", outFilePath)
			}

			// defer file closing
			defer jsonFile.Close()

			var teamData TeamData

			dataAsByte, readErr := io.ReadAll(jsonFile)

			if readErr != nil {
				LogErrorf(readErr, outFilePath)
			}

			//Deocding
			err := json.Unmarshal(dataAsByte, &teamData)

			//Deal with unmarshalling errors
			if err != nil {
				LogErrorf(err, "Error unmarshalling JSON data %v", string(dataAsByte))
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
	TeamNumber int `json:"Team"` // The team number
	// PitIdentifier string `json:"Pit"`     // The pit identifier, as seen on the pit map
	Scouter string `json:"Scouter"` // The person who did the pit scouting
	Notes   string `json:"Notes"`   // Other notes

	Weight  string `json:"Weight"`          //The Weight of the robot
	AutoNum string `json:"Number of Autos"` //Number of Autos
	Dynamic bool   `json:"Dyanamic Auto?"`  //Whether or not the team has dynamic autos

	Drivetrain          string    `json:"Drive Train"`                                   // The type of drivetrain the robot has
	GearRatio           string    `json:"Gear Ratio"`                                    //  The type of gearratio the robot has
	Coral               CoralData `json:"Coral Position"`                                //The position of the coral on the reef
	Algae               AlgaeData `json:"Algae Position"`                                //The position of the algae on the reef
	AlgaeGround         bool      `json:"Algae Ground Pickup"`                           //Whether the team is able to pick up from the ground
	AlgaeSource         bool      `json:"Algae Source Pickup"`                           //Whether the team is able to pick up from the source
	Cycle               int       `json:"Driver Years of Experience"`                    //How long the driver has been driving
	Experience          string    `json:"Cycle Time"`                                    //The team's average cycle time
	Teleop              int       `json:"Preferred Teleop"`                              //The preferred teleop???
	Endgame             int       `json:"Preferred Endgame"`                             //The preferred endgame
	Shallow             bool      `json:"Can Climb Shallow Cage"`                        //Whether it used the shallow climb
	Deep                bool      `json:"Can Climb Deep Cage"`                           //Whether it used the deep climb
	RobotTypeCompliment string    `json:"What Type of Robot Would Compliment You Best?"` //Question for pit scouting
	FavoritePart        string    `json:"Favorite Part of the Robot?"`                   //Question for pit scouting                               //Notes for other relevant information
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
type CoralData struct {
	L1 bool `json:"L1"` //L1 position of the reef
	L2 bool `json:"L2"` //L2 position of the reef
	L3 bool `json:"L3"` //L3 position of the reef
	L4 bool `json:"L4"` //L4 position of the reef
}
type AlgaeData struct {
	L2 bool `json:"A1"` //A1 position of the reef
	L3 bool `json:"A2"` //A2 position of the reef
}

// Parses through the file at the passed in location, returning a compiled PitScoutingData object and wether or not there were errors.
func ParsePitScout(file string) (PitScoutingData, bool) {

	path := filepath.Join(JsonInDirectory, file)

	// Open file
	jsonFile, fileErr := os.Open(path)

	// Handle any error opening the file
	if fileErr != nil {
		LogErrorf(fileErr, "Error opening JSON file %v", path)
		return PitScoutingData{}, true
	}

	// defer file closing
	defer jsonFile.Close()

	var pitData PitScoutingData

	dataAsByte, readErr := io.ReadAll(jsonFile)

	if readErr != nil {
		LogErrorf(readErr, "Error reading JSON file %v", path)
		return PitScoutingData{}, true
	}

	//Deocding
	err := json.Unmarshal(dataAsByte, &pitData)
	//Deal with unmarshalling errors
	if err != nil {
		LogErrorf(err, "Error unmarshalling JSON data %v", string(dataAsByte))
		return PitScoutingData{}, true
	}

	return pitData, false
}
