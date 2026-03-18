package internal

// Everything + the kitchen sink

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Simple wrapper for converting bool to string for replays
func GetReplayString(isReplay bool) string {
	if isReplay {
		return "replay"
	}
	return ""
}

func GetStyleString(tele TeleopData) string {
	var out string

	if tele.BotType != "Select" {
		out += tele.BotType + "; "
	}
	if tele.Playstyle != "Select" {
		out += tele.Playstyle
	}

	return out
}

// Utility to check if a given array of cycles is valid for writing
func cyclesAreValid(cycles []Cycle) bool {
	return len(cycles) > 0 && cycles[0].Type != "None"
}

// Gets the number of cycles out of an array of cycles while avoiding nulls, nones, and NaNs
func GetNumCycles(cycles []Cycle) int {
	if cyclesAreValid(cycles) {
		return len(cycles)
	}

	return 0
}

func GetTeleopCoverage(tele TeleField) string {
	// type TeleFieldV2 struct {
	// 	Bump   bool `json:"bump"`
	// 	Trench bool `json:"trench"`
	// }

	var out string

	if tele.Bump {
		out += "Over Bump; "
	} else if tele.Trench {
		out += "Under Trench"
	}

	return out
}

func GetCollection(data CollectionData) string {
	if data.CollectHP && data.CollectNeutral {
		return "HP & Neutral"
	} else if data.CollectHP {
		return "HP"
	} else if data.CollectNeutral {
		return "Neutral"
	}

	return ""
}

// Gets the average cycle time from an array of cycles, returning N/A if cycles are invalid
func GetAvgCycleTime(cycles []Cycle) any {
	if cyclesAreValid(cycles) {
		return cycles[len(cycles)-1].Time / float64(len(cycles))
	}
	return "N/A"
}

// Gets the average cycle time from an array of cycles, returning 0 if cycles are invalid.
// Used for more number-strenuous multi scouting
func GetAvgCycleTimeExclusive(cycles []Cycle) float64 {
	if cyclesAreValid(cycles) {
		return cycles[len(cycles)-1].Time / float64(len(cycles))
	}
	return 0
}

// Calculates the average accuracy of the passed in array of cycles, returning N/A if they are invalid.
func GetCycleAccuracy(cycles []Cycle) any {
	if cyclesAreValid(cycles) {
		totalAccuracy := 0.0
		for _, cycle := range cycles {
			totalAccuracy += cycle.Accuracy
		}
		return (float64(totalAccuracy) / float64(len(cycles))) * 100
	}
	return "N/A"
}

// Gets the accuracy of a robot during an autonomous period, returning N/A if 0 attempts were made
func GetAutoAccuracy(auto AutoData) any { // TODO: FIX FOR NEW
	attempts := auto.Scores + auto.Misses

	if attempts == 0 {
		return "N/A"
	}
	return (float64(auto.Scores) / float64(attempts)) * 100
}

func CompileNotes(team TeamData) string {
	var finalNote string = ""

	if team.Teleop.BotType != "" {
		finalNote += team.Teleop.Playstyle
	}
	if team.Teleop.Playstyle != "" {
		finalNote += team.Teleop.Playstyle
	}

	if team.Issues.LoseTrack {
		finalNote += "LOST TRACK; "
	}

	if team.Issues.Disconnect {
		finalNote += "DISCONNECTED; "
	}

	if team.Issues.EverBeached {
		finalNote += "WAS BEACHED; "
	}

	finalNote += team.Notes.Auto + "; "
	finalNote += team.Notes.Teleop + "; "
	finalNote += team.Notes.Perf + "; "
	finalNote += team.Notes.Events + "; "
	finalNote += team.Notes.Comments
	return finalNote
}

// Compiles Losing track, DCs, and notes into one string of notes.
// Used for multi-scouting only
func CompileNotes2(match MultiMatch, teams []TeamData) string {
	var finalNote string = ""
	var lostTrack bool = false
	var DC bool = false

	for _, entry := range teams {
		if entry.Issues.LoseTrack {
			lostTrack = true
		}

		if entry.Issues.Disconnect {
			DC = true
		}
	}

	if lostTrack {
		finalNote += "LOST TRACK; "
	}

	if DC {
		finalNote += "DISCONNECTED; "
	}

	finalNote += strings.Join(match.Notes, "; ")
	return finalNote
}

func TurnAutoFieldIntoAnAwesomeAndReadableString(autoField AutoField) string {
	var finalNote string = ""

	if autoField.Left {
		finalNote += "Left of field; "
	}

	if autoField.Mid {
		finalNote += "Middle of field; "
	}

	if autoField.Right {
		finalNote += "Right of field; "
	}

	if autoField.Top {
		finalNote += "Top of field; "
	}
	if autoField.Bump {
		finalNote += "Over Bump; "
	}

	if autoField.Trench {
		finalNote += "Under Trench; "
	}

	if autoField.DidntCross {
		finalNote += "Didn't Cross; "
	}

	if autoField.HP {
		finalNote += "HP Station; "
	}

	if autoField.Fuel {
		finalNote += "Fuel Station; "
	}

	return finalNote
}

// Returns if a file exists in Teamlists matching the passed in event key
func CheckForTeamLists(eventKey string) bool {
	_, err := os.Open(filepath.Join(CachedConfigs.TeamListsDirectory, eventKey))

	return err == nil
}

// Writes the teams attending an event to the matching file in TeamLists
func WriteTeamsToFile(configs GeneralConfigs) {
	runnable := exec.Command(configs.PythonDriver, "getTeamList.py", configs.TBAKey, configs.EventKey, configs.TeamListsDirectory)

	_, err := runnable.Output()

	if err != nil && !strings.Contains(err.Error(), "exit status 1") {
		LogErrorf(err, "Error executing command %v %v %v %v %v", configs.PythonDriver, "getTeamlist.py", configs.TBAKey, configs.EventKey, configs.TeamListsDirectory)
	}
}

// Reads the Teams from teamlists and stores them in memory
func StoreTeams() {
	pathToCurrEvent := filepath.Join(CachedConfigs.TeamListsDirectory, GetCurrentEvent())

	file, err := os.Open(pathToCurrEvent)

	if err != nil {
		LogErrorf(err, "Error opening %v", pathToCurrEvent)
	}

	resultBytes, readErr := io.ReadAll(file)
	resultStr := strings.Split(string(resultBytes), "\n")[1:]

	var resultInts []int
	for _, result := range resultStr {
		if result != "" {
			parsed, err := strconv.ParseInt(result, 10, 64)
			if err != nil {
				LogErrorf(err, "Error parsing %v as int", result)
			}
			resultInts = append(resultInts, int(parsed))
		}
	}

	if readErr != nil {
		LogErrorf(readErr, "Error reading %v", pathToCurrEvent)
	}

	Teams = resultInts
}

// Writes the schedule of an event to schedule/pkg.json
func WriteScheduleToFile(configs GeneralConfigs) {
	runnable := exec.Command(configs.PythonDriver, "getSchedule.py", configs.TBAKey, configs.EventKey, configs.RuntimeDirectory)

	_, err := runnable.Output()

	if err != nil && !strings.Contains(err.Error(), "exit status 1") {
		LogErrorf(err, "Error executing command %v %v %v", configs.PythonDriver, "getSchedule.py", configs.EventKey)
	}
}

// Writes all events for the current year to events.json
func WriteEventsToFile(configs GeneralConfigs) {
	runnable := exec.Command(configs.PythonDriver, "getAllEvents.py", configs.TBAKey)

	out, err := runnable.Output()

	if err != nil && !strings.Contains(err.Error(), "exit status 1") {
		LogErrorf(err, "Error executing command %v %v %v", configs.PythonDriver, "getAllEvents.py", configs.TBAKey)
	}

	if strings.Contains(string(out), "ERR") {
		LogMessagef("Error executing command %v %v %v; Investigate in python", configs.PythonDriver, "getAllEvents.py", configs.TBAKey)

	}
}

// Calculates the string from data pertaining to a driverstation
func GetDSString(isBlue bool, number uint) string {
	var builder string = ""

	if isBlue {
		builder += "blue"
	} else {
		builder += "red"
	}

	builder += fmt.Sprint(number)

	return builder
}

// turns the driverstation string into an int 1-5 representing its absolute number
func GetDSOffset(ds string) int {
	switch chooser := ds; chooser {
	case "red1":
		return 0
	case "red2":
		return 1
	case "red3":
		return 2
	case "blue1":
		return 3
	case "blue2":
		return 4
	case "blue3":
		return 5
	}

	return 0
}

// Gets the row an entry will write to from its Teamdata object
func GetRow(team TeamData) int { //TODO: Update so it doesn't rely on match # -Leon
	startRow := 2 + (team.Match.Number-1)*6
	dsString := GetDSString(team.DriverStation.IsBlue, uint(team.DriverStation.Number))
	dsOffset := GetDSOffset(dsString)

	startRow += uint(dsOffset)

	return int(startRow)
}

// Gets the row a pit scouting data should write to
func GetPitRow(team int) int {
	return slices.Index(Teams, team) + 1
}

// Getter for the current event key
func GetCurrentEvent() string {
	return CachedConfigs.EventKey
}

// Compares two slices for equality
func CompareSplits(first []string, second []string) bool {
	if len(first) != len(second) {
		return false
	}

	for i, element := range first {
		if element != second[i] {
			return false
		}
	}

	return true
}

// Gets all files matching the passed in pattern; Used to find all files of one entry when multi-scouting
func GetAllMatching(checkAgainst string) []string {
	var results []string
	splitAgainst := strings.Split(checkAgainst, "_")

	writtenJson, err := os.ReadDir(JsonWrittenDirectory)

	if err != nil {
		LogErrorf(err, "Error reading directory %v", JsonWrittenDirectory)
		return results
	}

	if len(writtenJson) > 0 {
		for _, jsonFile := range writtenJson {
			splitFile := strings.Split(jsonFile.Name(), "_")

			if len(splitFile) < 4 {
				continue
			}

			if CompareSplits(splitAgainst[:3], splitFile[:3]) {
				results = append(results, jsonFile.Name())
			}
		}
	}
	return results
}

// Gets the number of matches from json
func GetNumMatches() int {
	var result map[int]map[string][]int // pain

	jsonPath := filepath.Join(CachedConfigs.RuntimeDirectory, "json")
	file, err := os.Open(jsonPath)

	if err != nil {
		LogErrorf(err, "Error opening %v", jsonPath)
		return len(result)
	}

	decodeErr := json.NewDecoder(file).Decode(&result)
	if decodeErr != nil {
		LogErrorf(err, "Error Decoding %v", jsonPath)
		return len(result)
	}

	return len(result)
}

// Moves a file from an original path to a new one, returning wether or not it was successful
func MoveFile(originalPath string, newPath string) bool {
	oldLoc, openErr := os.Open(originalPath)

	if openErr != nil {
		LogErrorf(openErr, "Error opening %v", originalPath)
		return false
	}

	newLoc, openErr := OpenWithPermissions(newPath)
	if openErr != nil {
		LogErrorf(openErr, "Error creating %v", newPath)
		return false
	}

	defer newLoc.Close()

	_, copyErr := io.Copy(newLoc, oldLoc)

	if copyErr != nil {
		LogErrorf(copyErr, "Error copying %v to %v", originalPath, newPath)
		return false
	}

	if closeErr := oldLoc.Close(); closeErr != nil { //This is NOT a cause of returning false
		LogError(copyErr, "Error closing "+originalPath)
	}

	if removeErr := os.Remove(originalPath); removeErr != nil {
		LogError(removeErr, "Error removing "+originalPath)
		return false
	}

	return true
}
