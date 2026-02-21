package internal

// Utility for managing scouter schedules

import (
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
)

// Reference to the SQL scouting database
var scoutDB *sql.DB

// Opens the reference to the scouting database
func InitScoutDB() {
	dbPath := filepath.Join(CachedConfigs.RuntimeDirectory, "scout.db")
	dbRef, dbOpenErr := sql.Open(CachedConfigs.SqliteDriver, dbPath)

	scoutDB = dbRef

	if dbOpenErr != nil {
		LogErrorf(dbOpenErr, "Problem opening database %v", dbPath)
	}
}

// Struct containing the scouting range format encoded by the scheduling system
type ScoutRanges struct {
	Ranges [][3]int `json:"Ranges"` // A an array of arrays of ints of length 3, [dsoffset, starting, ending]
}

// Gets the schedule of one scouter
func RetrieveSingleScouter(name string, isUUID bool) string {
	var uuid string
	if isUUID {
		uuid = name
	} else {
		uuid, _ = GetUUID(name, true)
	}

	response := scoutDB.QueryRow("select schedule from individuals where uuid = ?", uuid)

	var ranges string

	scanErr := response.Scan(&ranges)
	if scanErr != nil && !errors.Is(scanErr, sql.ErrNoRows) {
		LogErrorf(scanErr, "Problem scanning response %v", response)
	}

	if ranges == "" {
		return `{"Ranges":[]}`
	} else {
		return ranges
	}
}

// Gets the schedule of one scouter, marshalled as a ScoutRanges object
func retrieveScouterAsObject(name string, isUUID bool) ScoutRanges {

	scheduleString := RetrieveSingleScouter(name, isUUID)

	var ranges ScoutRanges

	unmarshalErr := json.Unmarshal([]byte(scheduleString), &ranges)
	if unmarshalErr != nil {
		LogErrorf(unmarshalErr, "Problem Unmarshalling %v", []byte(scheduleString))
	}

	return ranges
}

// Adds a schedule update to an individual
func AddIndividualSchedule(name string, nameIsUUID bool, ranges ScoutRanges) {

	var uuid string
	if nameIsUUID {
		uuid = name
	} else {
		uuid, _ = GetUUID(name, true)
	}

	rangeBytes, marshalErr := json.Marshal(ranges)
	if marshalErr != nil {
		LogErrorf(marshalErr, "Problem marshalling %v", ranges)
	}

	rangeString := string(rangeBytes)

	if userInSchedule(scoutDB, uuid) { //If doesn't exist
		cachedRanges := retrieveScouterAsObject(name, nameIsUUID)

		var newRanges ScoutRanges
		newRanges.Ranges = append(cachedRanges.Ranges, ranges.Ranges...)

		newRangeBytes, err := json.Marshal(newRanges)
		if err != nil {
			LogErrorf(err, "Problem marshalling %v", ranges)
		}

		rangeString = string(newRangeBytes)

		_, resultErr := scoutDB.Exec("update individuals set schedule = ? where uuid = ?", rangeString, uuid)
		if resultErr != nil {
			LogErrorf(resultErr, "Problem executing sql command %v with args %v", "update individuals set ranges = ? where uuid = ?", []any{rangeString, uuid})
		}

	} else {
		user := UUIDToUser(uuid)
		_, resultErr := scoutDB.Exec("insert into individuals values(?, ?, ?)", uuid, user, rangeString)
		if resultErr != nil {
			LogErrorf(resultErr, "Problem executing sql command %v with args %v", "insert into individuals values(?, ?, ?)", []any{uuid, user, rangeString})
		}
	}

}

// Returns if an individual has any schedule entries
func userInSchedule(database *sql.DB, uuid string) bool {
	result := database.QueryRow("select count(1) from individuals where uuid = ?", uuid)

	var resultstore int
	err := result.Scan(&resultstore)

	if err != nil {
		LogErrorf(err, "Problem scanning response %v", result)
	}

	return resultstore == 1
}

// Wipes the json file
func WipeSchedule() {
	schedPath := filepath.Join(CachedConfigs.RuntimeDirectory, "json")
	file, openErr := OpenWithPermissions(schedPath)

	if openErr != nil {
		LogErrorf(openErr, "Problem opening %v", schedPath)
	}

	_, writeErr := file.WriteString("{}")
	if writeErr != nil {
		LogErrorf(writeErr, "Problem resetting %v", schedPath)
	} else {
		LogMessage("Successfully wiped json")
	}

	file.Close()
}
