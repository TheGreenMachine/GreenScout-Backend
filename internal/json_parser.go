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

type TeamData struct {
	TeamNumber    uint64            `json:"team"`
	Match         MatchInfo         `json:"match"`
	Scouter       string            `json:"scouter"`
	DriverStation DriverStationData `json:"driverStation"`
	Cycles        []Cycle           `json:"cycles"` // The cycle data

	Auto    AutoData    `json:"auto"`
	Teleop  TeleopData  `json:"teleop"`
	Endgame EndgameData `json:"endgame"`
	Issues  IssuesData  `json:"issues"`
	Notes   NotesData   `json:"notes"`

	Rescouting  bool `json:"rescouting"`
	Prescouting bool `json:"prescouting"`
}

type AutoData struct {
	CanAuto  bool `json:"canAuto"`
	HangAuto bool `json:"hangAuto"`
	Scores   int  `json:"scores"`
	Misses   int  `json:"misses"`
	Ejects   int  `json:"ejects"`
	WonAuto  bool `json:"won"`

	Accuracy AutoAccuracy `json:"accuracy"`
	Field    AutoField    `json:"field"`
}

type AutoAccuracy struct {
	HPAccuracy    int `json:"hpAccuracy"`
	RobotAccuracy int `json:"robotAccuracy"`
}

type AutoField struct {
	Left       bool `json:"left"`
	Right      bool `json:"right"`
	Mid        bool `json:"mid"`
	Top        bool `json:"top"`
	Bump       bool `json:"bump"`
	Trench     bool `json:"trench"`
	DidntCross bool `json:"didntCross"`
	HP         bool `json:"hp"`
	Fuel       bool `json:"fuel"`
}

type TeleopData struct {
	Collection CollectionData `json:"collection"`
	Field      TeleField      `json:"field"`
	BotType    string         `json:"botType"`
	Playstyle  string         `json:"playstyle"`
}

type CollectionData struct {
	CollectNeutral bool   `json:"collectNeutral"`
	CollectHP      bool   `json:"collectHp"`
	FuelCapacity   string `json:"fuelCapacity"`
}

type TeleField struct {
	Bump   bool `json:"bump"`
	Trench bool `json:"trench"`
}

type EndgameData struct {
	Park         string  `json:"park"`
	ClimbTimer   float64 `json:"climbTimer"`
	EndgameShoot bool    `json:"endgameShoot"`
}

type IssuesData struct {
	Disconnect  bool `json:"disconnect"`
	LoseTrack   bool `json:"loseTrack"`
	EverBeached bool `json:"everBeached"`
}

type NotesData struct {
	Perf     string `json:"perfNotes"`
	Events   string `json:"eventsNotes"`
	Comments string `json:"commentsNotes"`
	Teleop   string `json:"teleNotes"`
	Auto     string `json:"autoNotes"`
}

// Basic info about the driver station
type DriverStationData struct {
	IsBlue bool `json:"isBlue"` // If it is blue
	Number int  `json:"number"` // The driverstation number (1-3)
}

// Basic info about the match
type MatchInfo struct {
	Number   uint `json:"number"`   // The match number
	IsReplay bool `json:"isReplay"` // If it is a replay
}

// One cycle
type Cycle struct {
	Time     float64 `json:"time"`     // The time taken
	Type     string  `json:"type"`     // The type of cycle
	Accuracy float64 `json:"accuracy"` // The accuracy of the cycle. Will also be drove and shot for shuttles
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
