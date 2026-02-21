package internal

// Utilities for interacting with users.db

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"slices"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// The reference to users.db
var userDB *sql.DB

// Initializes users.db and stores the reference to memory
func InitUserDB() {
	dbPath := filepath.Join(CachedConfigs.PathToDatabases, "users.db")
	dbRef, dbOpenErr := sql.Open(CachedConfigs.SqliteDriver, dbPath)

	userDB = dbRef

	if dbOpenErr != nil {
		FatalError(dbOpenErr, "Problem opening database "+dbPath)
	}
}

// Creates a new user
func NewUser(username string, uuid string) {
	badgeBytes, marshalError := json.Marshal(emptyBadges())

	if marshalError != nil {
		LogError(marshalError, "Problem marshalling empty badge JSON")
	}

	//The only reason most of these columns don't have default values is that sqlite doesn't let you alter column default values and I don't feel like deleting and remaking every column
	_, err := userDB.Exec("insert into users values(?,?,?,?,?,?,?, 0, 0, ?, 0)", uuid, username, username, nil, string(badgeBytes), 0, DefaultPfpPath, "[]")

	if err != nil {
		LogErrorf(err, "Problem creating new user with args: %v, %v, %v, %v, %v, %v, %v", uuid, username, username, "nil", badgeBytes, 0, DefaultPfpPath)
	}
}

// Returns if a user exists
func userExists(username string) bool {
	// Basically, count up for every time there is a user with this username
	result := userDB.QueryRow("select count(1) from users where username = ?", username)

	var resultstore int
	scanErr := result.Scan(&resultstore)

	if scanErr != nil {
		LogError(scanErr, "Problem scanning response to sql query SELECT COUNT(1) FROM users WHERE username = ? with arg: "+username)
	}

	// If we ever get more than 1, something is horribly wrong.s
	return resultstore == 1
}

// Returns the uuid of a user. If the user does not exist, it will check the createIfNot boolean. If this is true, it will create
// a new user and return its uuid. If not, it will return an empty string and false
func GetUUID(username string, createIfNot bool) (string, bool) {
	userExists := userExists(username)

	if (!createIfNot) && !userExists {
		return "", false
	}

	if !userExists {
		NewUser(username, "") //Empty UUID for assignment later
	}

	var userId string
	result := userDB.QueryRow("select uuid from users where username = ?", username)
	scanErr := result.Scan(&userId)
	if scanErr != nil {
		LogError(scanErr, "Problem scanning response to sql query SELECT uuid FROM users WHERE username = ? with arg: "+username)
	}

	if userId == "" {
		newId := uuid.New()
		_, err := userDB.Exec("update users set uuid = ? where username = ?", newId, username)

		if err != nil {
			LogErrorf(err, "Problem executing sql query UPDATE users SET uuid = ? WHERE username = ? with args: %v, %v", newId, username)
		}

		userId = newId.String()
	}
	return userId, true
}

// Converts a uuid to a username
func UUIDToUser(uuid string) string {
	result := userDB.QueryRow("select username from users where uuid = ?", uuid)
	var resultStore string
	err := result.Scan(&resultStore)

	if err != nil {
		LogErrorf(err, "Problem scanning results of sql query SELECT username FROM users WHERE uuid = ? with arg: %v", uuid)
	}

	return resultStore
}

// A user
type User struct {
	Name string // The username
	UUID string // The uuid
}

// Returns all users
func GetAllUsers() []User {
	result, err := userDB.Query("select username, uuid from users")

	if err != nil {
		LogError(err, "Problem executing sql query SELECT username, uuid FROM users")
	}

	var users []User

	for result.Next() {
		var name string
		var uuid string
		scanErr := result.Scan(&name, &uuid)

		if scanErr != nil {
			LogError(scanErr, "Problem scanning results of sql query SELECT username, uuid FROM users")
		}

		if name != "" {
			users = append(users, User{Name: name, UUID: uuid})
		}
	}

	return users
}

// A badge
type Badge struct {
	ID          string // The badge name
	Description string // The badge description
}

type Accolade string

// An accolade, as well as if the frontend has been notified of it.
type AccoladeData struct {
	Accolade Accolade
	Notified bool
}

// User information
type UserInfo struct {
	Username    string         // The username
	DisplayName string         // The display name
	Accolades   []AccoladeData // The leaderboard-invisible achievements and silent badges
	Badges      []Badge        // The leaderboard-visible badges
	Score       int            // The score
	LifeScore   int            // The lifetime score
	HighScore   int            // The high score
	Color       LBColor        // The leaderboard color
	Pfp         string         // The relative path to the profile picture
}

// User information to be served for admins to edit
type UserInfoForAdmins struct {
	Username    string  // The username
	DisplayName string  // The displayname
	UUID        string  // The uuid
	Color       LBColor // The leaderboard color
	Badges      []Badge // The badges
}

// Returns user info for admins to edit
func GetAdminUserInfo(uuid string) UserInfoForAdmins {
	if uuid == "[[CURRENTLY NO ACTIVE USER IS SELECTED AT THIS MOMENT]]" {
		return UserInfoForAdmins{}
	}

	username := UUIDToUser(uuid)

	var displayName string
	var color LBColor
	var badges []Badge

	displayName = GetDisplayName(uuid)
	color = getLeaderboardColor(uuid)
	badges = GetBadges(uuid)

	userInfo := UserInfoForAdmins{
		Username:    username,
		UUID:        uuid,
		DisplayName: displayName,
		Color:       color,
		Badges:      badges,
	}
	return userInfo
}

// Returns the user information of a given username
func GetUserInfo(username string) UserInfo {
	uuid, exists := GetUUID(username, false)

	var displayName string
	var accolades []AccoladeData
	var badges []Badge
	var score int
	var lifeScore int
	var highscore int
	var color LBColor

	var pfp string

	if exists { // This could 100% be made more efficient, but not my problem!
		displayName = GetDisplayName(uuid)
		badges = GetBadges(uuid)
		accolades = GetAccolades(uuid)
		score = getScore(uuid)
		lifeScore = getLifeScore(uuid)
		highscore = getHighScore(uuid)
		color = getLeaderboardColor(uuid)
		pfp = getPfp(uuid)
	} else {
		displayName = "User does not exist"
		badges = emptyBadges()
		accolades = emptyAccolades()
		score = -1
		lifeScore = -1
		highscore = -1
		color = 0
		pfp = DefaultPfpPath
	}

	userInfo := UserInfo{
		Username:    username,
		DisplayName: displayName,
		Badges:      badges,
		Accolades:   accolades,
		Score:       score,
		LifeScore:   lifeScore,
		HighScore:   highscore,
		Color:       color,
		Pfp:         filepath.Join(CachedConfigs.PfpDirectory, pfp),
	}
	return userInfo
}

// Gets the display name from a uuid
func GetDisplayName(uuid string) string {
	var displayName string
	response := userDB.QueryRow("select displayname from users where uuid = ?", uuid)
	scanErr := response.Scan(&displayName)
	if scanErr != nil {
		LogError(scanErr, "Problem scanning results of sql query SELECT displayname FROM users WHERE uuid = ? with arg: "+uuid)
	}

	return displayName
}

// Gets the badges from a uuid
func GetBadges(uuid string) []Badge {
	var Badges []Badge
	var BadgesMarshalled string
	response := userDB.QueryRow("select badges from users where uuid = ?", uuid)
	scanErr := response.Scan(&BadgesMarshalled)
	if scanErr != nil {
		LogError(scanErr, "Problem scanning results of sql query SELECT badges FROM users WHERE uuid = ? with arg: "+uuid)
	}
	// I am aware of how awful converting []byte -> string -> []byte is but i've had problems storing byte arrays with sqlite. If you guys want to switch to postgres it would solve it but that's a fairly steep learning curve
	unmarshalErr := json.Unmarshal([]byte(BadgesMarshalled), &Badges)
	if unmarshalErr != nil {
		LogErrorf(unmarshalErr, "Problem unmarshalling %v", BadgesMarshalled)
	}

	return Badges
}

// Generates an empty array of badges
func emptyBadges() []Badge {
	return []Badge{}
}

// Generates an empty array of accolades
func emptyAccolades() []AccoladeData {
	return []AccoladeData{}
}

// Sets the leaderboard color of a uuid
func SetColor(uuid string, color LBColor) {
	_, execErr := userDB.Exec("update users set color = ? where uuid = ?", color, uuid)
	if execErr != nil {
		LogErrorf(execErr, "Problem executing sql query UPDATE users SET color = ? WHERE uuid = ? with args: %v, %v", color, uuid)
	}
}

// Sets the display name of a given user
func SetDisplayName(username string, displayName string) {
	uuid, _ := GetUUID(username, true)

	_, execErr := userDB.Exec("update users set displayname = ? where uuid = ?", displayName, uuid)

	if execErr != nil {
		LogErrorf(execErr, "Problem executing sql query UPDATE users SET displayname = ? WHERE uuid = ? with args: %v, %v", displayName, uuid)
	}
}

// Adds an accolade to a given user
func AddAccolade(uuid string, accolade Accolade, frontendAchievement bool) {
	existingAccolades := GetAccolades(uuid)
	existingAccoladeNames := ExtractNames(existingAccolades)

	// only append if it isn't already present
	if !slices.Contains(existingAccoladeNames, accolade) {
		existingAccolades = append(existingAccolades, AccoladeData{Accolade: accolade, Notified: frontendAchievement})
	}

	accBytes, marshalErr := json.Marshal(existingAccolades)
	if marshalErr != nil {
		LogErrorf(marshalErr, "Problem marshalling %v", existingAccolades)
	}

	_, execErr := userDB.Exec("update users set accolades = ? where uuid = ?", string(accBytes), uuid)
	if execErr != nil {
		LogErrorf(execErr, "Problem executing sql query UPDATE users SET accolades = ? WHERE uuid = ? with args: %v, %v", accBytes, uuid)
	}
}

// Sets the accolades of a given user to a passed in array of Accolade Data
func SetAccolades(uuid string, accolades []AccoladeData) {
	accBytes, marshalErr := json.Marshal(accolades)
	if marshalErr != nil {
		LogErrorf(marshalErr, "Problem marshalling %v", accolades)
	}

	_, execErr := userDB.Exec("update users set accolades = ? where uuid = ?", string(accBytes), uuid)
	if execErr != nil {
		LogErrorf(execErr, "Problem executing sql query UPDATE users SET accolades = ? WHERE uuid = ? with args: %v, %v", accBytes, uuid)
	}
}

// Adds a badge to a given user
func AddBadge(uuid string, badge Badge) {
	existingBadges := GetBadges(uuid)

	var toAppend = true
	// Go through, check if the ID is already present, and update descriptions if it is
	for i := range existingBadges {
		if existingBadges[i].ID == badge.ID {
			toAppend = false

			if existingBadges[i].Description != badge.Description {
				existingBadges[i].Description = badge.Description
			}

			break
		}
	}

	if toAppend {
		existingBadges = append(existingBadges, badge)
	}

	badgesBytes, marshalErr := json.Marshal(existingBadges)
	if marshalErr != nil {
		LogErrorf(marshalErr, "Problem marshalling %v", existingBadges)
	}

	_, execErr := userDB.Exec("update users set badges = ? where uuid = ?", string(badgesBytes), uuid)
	if execErr != nil {
		LogErrorf(execErr, "Problem executing sql query UPDATE users SET badges = ? WHERE uuid = ? with args: %v, %v", badgesBytes, uuid)
	}

}

func SetBadges(uuid string, badges []Badge) {
	badgesBytes, marshalErr := json.Marshal(badges)
	if marshalErr != nil {
		LogErrorf(marshalErr, "Problem marshalling %v", badges)
	}

	_, execErr := userDB.Exec("update users set badges = ? where uuid = ?", string(badgesBytes), uuid)
	if execErr != nil {
		LogErrorf(execErr, "Problem executing sql query UPDATE users SET badges = ? WHERE uuid = ? with args: %v, %v", badgesBytes, uuid)
	}
}

// Gets the score from a given user
func getScore(uuid string) int {
	var score int
	response := userDB.QueryRow("select score from users where uuid = ?", uuid)
	scanErr := response.Scan(&score)
	if scanErr != nil {
		LogError(scanErr, "Problem scanning response to sql query SELECT score FROM users WHERE uuid = ? with arg: "+uuid)
	}

	return score
}

// Gets the lifetime score of a given user
func getLifeScore(uuid string) int {
	var score int
	response := userDB.QueryRow("select lifescore from users where uuid = ?", uuid)
	scanErr := response.Scan(&score)
	if scanErr != nil {
		LogError(scanErr, "Problem scanning response to sql query SELECT lifescore FROM users WHERE uuid = ? with arg: "+uuid)
	}

	return score
}

// Gets the high score of a given user
func getHighScore(uuid string) int {
	var highscore int
	response := userDB.QueryRow("select highscore from users where uuid = ?", uuid)
	scanErr := response.Scan(&highscore)
	if scanErr != nil {
		LogError(scanErr, "Problem scanning response to sql query SELECT highscore FROM users WHERE uuid = ? with arg: "+uuid)
	}

	return highscore
}

type LBColor int

// Leaderboard color enum
const (
	Default LBColor = 0
	Green   LBColor = 1
	Gold    LBColor = 2
)

// Gets the leaderboard color of a given user
func getLeaderboardColor(uuid string) LBColor {
	var color LBColor
	response := userDB.QueryRow("select color from users where uuid = ?", uuid)
	scanErr := response.Scan(&color)
	if scanErr != nil {
		LogError(scanErr, "Problem scanning response to sql query SELECT color FROM users WHERE uuid = ? with arg: "+uuid)
	}

	return color
}

// Gets the relative path of a given user's profile picture
func getPfp(uuid string) string {
	var pfp string
	response := userDB.QueryRow("select pfp from users where uuid = ?", uuid)
	scanErr := response.Scan(&pfp)
	if scanErr != nil {
		LogError(scanErr, "Problem scanning response to sql query SELECT pfp FROM users WHERE uuid = ? with arg: "+uuid)
	}

	return pfp
}

// Sets a given user's path to profile picture
func SetPfp(username string, pfp string) {
	uuid, _ := GetUUID(username, true)

	_, execErr := userDB.Exec("update users set pfp = ? where uuid = ?", pfp, uuid)

	if execErr != nil {
		LogErrorf(execErr, "Problem executing sql query UPDATE users SET pfp = ? WHERE uuid = ? with args: %v, %v", pfp, uuid)
	}
}
