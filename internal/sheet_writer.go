package internal

// Utilites for accessing the google sheets API

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	yaml "sigs.k8s.io/yaml/goyaml.v2"
)

// Early methods (setup) are from google's quickstart, so I didn't change much about them

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := SheetsTokenFile
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		client, _ := google.DefaultClient(context.Background(), sheets.SpreadsheetsScope)
		if client != nil {
			return client
		}

		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	LogMessagef("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		FatalError(err, "Unable to read authorization code: ")
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		FatalError(err, "Unable to retrieve token from web: ")
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	LogMessagef("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		FatalError(err, "Unable to cache oauth token: ")
	}
	defer f.Close()
	encodeErr := json.NewEncoder(f).Encode(token)
	if encodeErr != nil {
		FatalError(encodeErr, "Unable to encode token to file")
	}
}

// The spreadsheet ID, held in memory
var SpreadsheetId string

// The service (api instance), held in memory
var Srv *sheets.Service

// Sets up the sheets API based on the credentials.json and token.json
func SetupSheetsAPI(creds []byte) {
	ctx := context.Background()

	config, err := google.ConfigFromJSON(creds, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		FatalError(err, "Unable to parse client secret file to config: %v")
	}
	client := getClient(config)

	Srv, err = sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		FatalError(err, "Unable to retrieve Sheets client: %v")
	}
	LogMessagef("Client retrieved for: %v", Srv.UserAgent)

	SpreadsheetId = CachedConfigs.SpreadSheetID
}

// Writes team data from multi-scouting to a specified line
func WriteMultiScoutedTeamDataToLine(matchdata MultiMatch, row int, sources []TeamData) bool { // TODO: FIX FOR NEW
	// troughTendency, L2Tendency, L3Tendency, L4Tendency, processorTendency, netTendency, knockTendency, shuttleTendency := GetCycleTendencies(matchdata.CycleData.AllCycles)
	// troughAccuracy, L2Accuracy, L3Accuracy, L4Accuracy, processorAccuracy, netAccuracy, knockAccuracy, shuttleAccuracy := GetCycleAccuracies(matchdata.CycleData.AllCycles)

	// This is ONE ROW. Each value is a cell in that row.
	valuesToWrite := []interface{}{
		GetDSString(matchdata.DriverStation.IsBlue, uint(matchdata.DriverStation.Number)),
		matchdata.Match,
		matchdata.TeamNumber,
		matchdata.CycleData.AvgCycleTime,
		matchdata.CycleData.NumCycles,
		TurnAutoFieldIntoAnAwesomeAndReadableString(matchdata.Auto.Field),
		matchdata.Auto.CanAuto,            // Had Auto
		matchdata.Auto.HangAuto,           // Had Auto
		matchdata.Auto.WonAuto,            // Had Auto
		matchdata.Auto.Scores,             // Scores in auto
		GetAutoAccuracy(matchdata.Auto),   // Auto accuracy
		matchdata.Auto.Ejects,             // Auto shuttles
		matchdata.Parked,                  // Parked
		CompileNotes2(matchdata, sources), // Notes + Penalties + DC + Lost track
	}

	var vr sheets.ValueRange

	vr.Values = append(vr.Values, valuesToWrite)

	writeRange := fmt.Sprintf("RawData!B%v", row)

	_, err := Srv.Spreadsheets.Values.Update(SpreadsheetId, writeRange, &vr).ValueInputOption("RAW").Do()

	if err != nil {
		LogError(err, "Unable to write data to sheet")
		return false
	}
	return true
}

// Writes data from a single-scouted match to a line
func WriteTeamDataToLine(teamData TeamData, row int) bool { // TODO: FIX FOR NEW
	// This is ONE ROW. Each value is a cell in that row.
	valuesToWrite := []interface{}{
		GetDSString(teamData.DriverStation.IsBlue, uint(teamData.DriverStation.Number)),
		teamData.Match.Number,            // Match Number
		teamData.TeamNumber,              // Team Number
		GetAvgCycleTime(teamData.Cycles), // Avg cycle time
		GetNumCycles(teamData.Cycles),    // Num Cycles
		GetCollection(teamData.Teleop.Collection),
		TurnAutoFieldIntoAnAwesomeAndReadableString(teamData.Auto.Field),
		teamData.Auto.CanAuto,                                     // Had Auto
		teamData.Auto.HangAuto,                                    // Had Hanging Auto
		teamData.Auto.WonAuto,                                     // Won Auto
		teamData.Auto.Scores,                                      // Scores in auto
		GetAutoAccuracy(teamData.Auto),                            // Auto accuracy
		fmt.Sprintf("%v%%", teamData.Auto.Accuracy.HPAccuracy),    // Accuracy of Human
		fmt.Sprintf("%v%%", teamData.Auto.Accuracy.RobotAccuracy), // Accuracy of Robot
		teamData.Auto.Ejects,                                      // Auto shuttles
		teamData.Endgame.ClimbTimer,                               // Climb Time
		teamData.Endgame.Park,                                     // Parked
		GetStyleString(teamData.Teleop),
		CompileNotes(teamData), // Notes + Penalties + DC + Lost track
	}

	var vr sheets.ValueRange

	vr.Values = append(vr.Values, valuesToWrite)

	writeRange := fmt.Sprintf("RawData!B%v", row)

	_, err := Srv.Spreadsheets.Values.Append(SpreadsheetId, writeRange, &vr).ValueInputOption("RAW").InsertDataOption("INSERT_ROWS").Do()

	if err != nil {
		LogError(err, "Unable to write data to sheet")
		return false
	}

	return true

}

// Wrapper around sheets' batch update.
func BatchUpdate(dataset [][]interface{}, writeRange string) {
	rb := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
	}

	rb.Data = append(rb.Data, &sheets.ValueRange{
		Range:  writeRange,
		Values: dataset,
	})

	_, err := Srv.Spreadsheets.Values.BatchUpdate(SpreadsheetId, rb).Do()

	if err != nil {
		LogError(err, "Unable to write data to sheet")
	}
}

// Fills sheet with all matches from that event.
func FillMatches(startMatch int, endMatch int) {
	if !(math.Abs(float64(endMatch)-float64(startMatch)) >= 50) {

		matchTracker := 2 + (startMatch-1)*6

		for i := startMatch; i <= endMatch; i++ {

			perMatchInterface := [][]interface{}{ // 6 numbers, all same
				{i}, {i}, {i}, {i}, {i}, {i},
			}

			BatchUpdate(perMatchInterface, fmt.Sprintf("RawData!A%v:A%v", matchTracker, matchTracker+6))
			matchTracker += 6
		}
	} else {
		LogMessage("Input matches with a delta under 50!")
	}
}

// Updates the ID of the sheet to be used, in memory and yaml.
func UpdateSheetID(newSheet string) string {
	if IsSheetValid(newSheet) {
		CachedConfigs.SpreadSheetID = newSheet

		configFile, openErr := OpenWithPermissions(ConfigFilePath)
		if openErr != nil {
			LogErrorf(openErr, "Problem opening %v", ConfigFilePath)
			return "There was a problem updating the sheet ID"
		}

		defer configFile.Close()

		encodeErr := yaml.NewEncoder(configFile).Encode(&CachedConfigs)

		if encodeErr != nil {
			LogErrorf(encodeErr, "Problem encoding %v", CachedConfigs)
			return "There was a problem updating the sheet ID"
		}

		return "Successfully updated sheet ID to " + newSheet
	}
	return "Sheet ID " + newSheet + " is invalid!"

}

// Tries to read the top-left cell of the RawData tab, returning if it can.
func IsSheetValid(id string) bool {
	spreadsheetId := id
	readRange := "RawData!A1:1"
	_, err := Srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	return err == nil
}

// Adds conditinoal formatting to the raw data tab.
// This consists of two sinusoidal functions that ensure 3-red 3-blue coloring.
func WriteConditionalFormatting() {
	tabs, err := Srv.Spreadsheets.Get(SpreadsheetId).Do()
	if err != nil {
		LogError(err, "Failed to get tabs")
		return
	}

	var sheetID int64

	for _, sheet := range tabs.Sheets {
		if sheet.Properties.Title == "RawData" {
			sheetID = sheet.Properties.SheetId
			break
		}
	}

	_, sheetErr := Srv.Spreadsheets.BatchUpdate(
		SpreadsheetId,
		&sheets.BatchUpdateSpreadsheetRequest{

			Requests: []*sheets.Request{
				{
					AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
						Index: 0,
						Rule: &sheets.ConditionalFormatRule{
							BooleanRule: &sheets.BooleanRule{
								Condition: &sheets.BooleanCondition{
									Type: "CUSTOM_FORMULA",
									Values: []*sheets.ConditionValue{
										{UserEnteredValue: "=(SIN(((PI() /3)) * (ROW()-1) -0.5)) > 0"},
									},
								},
								Format: &sheets.CellFormat{
									BackgroundColor: &sheets.Color{
										Red:   1,
										Alpha: 1, // https://steamuserimages-a.akamaihd.net/ugc/2040738890178501955/DB9342C662AFAF139B605B3B6EBF593ADF42550E/?imw=637&imh=358&ima=fit&impolicy=Letterbox&imcolor=%23000000
									},
								},
							},
							Ranges: []*sheets.GridRange{
								{
									SheetId:          sheetID,
									StartRowIndex:    1,
									StartColumnIndex: 0,
									EndColumnIndex:   1,
								},
							},
						},
					},
				},
				{
					AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
						Index: 1,
						Rule: &sheets.ConditionalFormatRule{
							BooleanRule: &sheets.BooleanRule{
								Condition: &sheets.BooleanCondition{
									Type: "CUSTOM_FORMULA",
									Values: []*sheets.ConditionValue{
										{UserEnteredValue: "=(SIN(((PI() /3)) * (ROW()-1) -0.5)) < 0"},
									},
								},
								Format: &sheets.CellFormat{
									BackgroundColor: &sheets.Color{
										Red:   164.0 / 255.0,
										Green: 194.0 / 255.0,
										Blue:  244.0 / 255.0,
										Alpha: 1, // https://i1.sndcdn.com/artworks-JyCZdFbdVSMdUUjr-driMCA-t500x500.jpg
										// https://media1.giphy.com/media/v1.Y2lkPTc5MGI3NjExdGQ0bGFyeHN5MmlidTNlMWVrYnRlZzNqdXdicXoxd2E0Y3F1NWVibiZlcD12MV9pbnRlcm5hbF9naWZfYnlfaWQmY3Q9Zw/ConhaVuI4urdeQ1wSk/giphy.gif
									},
								},
							},
							Ranges: []*sheets.GridRange{
								{
									SheetId:          sheetID,
									StartRowIndex:    1,
									StartColumnIndex: 0,
									EndColumnIndex:   1,
								},
							},
						},
					},
				},
			},
		},
	).Do()

	if sheetErr != nil {
		LogError(sheetErr, "Problem adding conditionall formatting.")
	}
}

// Writes data from pit scouting to a line
func WritePitDataToLine(pitData PitScoutingData, row int) bool {

	// This is ONE ROW. Each value is a cell in that row.
	valuesToWrite := []interface{}{
		pitData.TeamNumber,          //Team Number
		pitData.Scouter,             //Person/people who pit scouted
		pitData.Weight,              //The weight of the robot
		pitData.AutoNum,             //The number of autos they have
		pitData.Dynamic,             //Whether they have dynamic autos
		pitData.Drivetrain,          //The type of drivetrain
		pitData.GearRatio,           //The GearRatio on the top of my head
		pitData.Coral,               //The position(s) their robot is able to score
		pitData.Algae,               //The position(s) their robot is able to score
		pitData.AlgaeGround,         //Whether it can collect algae from the ground
		pitData.AlgaeSource,         //Whether it can collect algae from the source
		pitData.Cycle,               //Their cycle time
		pitData.Experience,          //The driver's experience
		pitData.Teleop,              //The strategy for teleop??
		pitData.Endgame,             //The strategy for endgame
		pitData.Shallow,             //Whether it can shallow climb
		pitData.Deep,                //Whether it can deep climb
		pitData.RobotTypeCompliment, //What part of the robot compliments you?
		pitData.FavoritePart,        //Favortite part of the robot
		pitData.Notes,               //Other Notes

	}

	var vr sheets.ValueRange

	vr.Values = append(vr.Values, valuesToWrite)

	writeRange := fmt.Sprintf("PitScouting!B%v", row)

	_, err := Srv.Spreadsheets.Values.Append(SpreadsheetId, writeRange, &vr).ValueInputOption("RAW").InsertDataOption("INSERT_ROWS").Do()

	if err != nil {
		LogError(err, "Unable to write data to sheet")
		return false
	}

	return true

}

// Writes data from a prescouted match to a line
func WritePrescoutDataToLine(teamData TeamData, row int) bool { // TODO: FIX FOR NEW

	// This is ONE ROW. Each value is a cell in that row.
	valuesToWrite := []interface{}{
		GetDSString(teamData.DriverStation.IsBlue, uint(teamData.DriverStation.Number)),
		teamData.TeamNumber,              // Team Number
		GetAvgCycleTime(teamData.Cycles), // Avg cycle time
		GetNumCycles(teamData.Cycles),    // Num Cycles

		teamData.Auto.CanAuto,          // Had Auto
		teamData.Auto.Scores,           // Scores in auto
		GetAutoAccuracy(teamData.Auto), // Auto accuracy
		teamData.Auto.Ejects,           // Auto shuttles
		teamData.Endgame.ClimbTimer,    // Climb Time
		// GetParkStatus(teamData.Endgame),           // Parked
		CompileNotes(teamData), // Notes + Penalties + DC + Lost track
	}

	var vr sheets.ValueRange

	vr.Values = append(vr.Values, valuesToWrite)

	writeRange := fmt.Sprintf("Prescouting!B%v", row)

	_, err := Srv.Spreadsheets.Values.Update(SpreadsheetId, writeRange, &vr).ValueInputOption("RAW").Do()

	if err != nil {
		LogError(err, "Unable to write data to sheet")
		return false
	}

	return true

}
