package internal

// Utilities for interacting with users.db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// The reference to users.db
var userDB *sql.DB

type UserColumn struct {
	columnName   string         // The name of the SQL column
	valueType    string         // The type of the column (TEXT, INT, etc...)
	defaultValue sql.NullString // The value that will be in row if not replaced
	notNull      bool           // Prevents a column from being null
	primaryKey   bool           // If the column should be used to look up a row. Also enforces uniqueness
	unique       bool           // Each row has to have a different value in the column
}

// Initializes users.db and stores the reference to memory
func InitUserDB() {
	dbPath := filepath.Join(CachedConfigs.PathToDatabases, "users.db")
	_, err := os.Stat(dbPath)
	dbMissing := err != nil && errors.Is(err, os.ErrNotExist)

	dbRef, err := sql.Open(CachedConfigs.SqliteDriver, dbPath)
	if err != nil {
		FatalError(err, "Problem opening database "+dbPath)
	}
	userDB = dbRef

	// when changing this, please make sure it you modify NewUser() as new account creation may fail for whatever reason
	schema := []UserColumn{
		{columnName: "uuid", valueType: "TEXT", unique: true, primaryKey: true},
		{columnName: "username", valueType: "TEXT", unique: true},
		{columnName: "displayname", valueType: "TEXT"},
		{columnName: "certificate", valueType: "TEXT"},
		{columnName: "badges", valueType: "TEXT[]"},
		{columnName: "score", valueType: "INT"},
		{columnName: "pfp", valueType: "TEXT", defaultValue: sql.NullString{String: "'" + DefaultPfpPath + "'", Valid: true}},
		{columnName: "lifescore", valueType: "INT"},
		{columnName: "highscore", valueType: "INT"},
		{columnName: "accolades", valueType: "TEXT", defaultValue: sql.NullString{String: "'[]'", Valid: true}},
		{columnName: "color", valueType: "INT"},
		{columnName: "theme", valueType: "TEXT", defaultValue: sql.NullString{String: "'Light'", Valid: true}},
	}

	if dbMissing {
		LogMessage("Creating new user database. Populating...")
		err = populateNewDB(schema)
		if err != nil {
			FatalError(err, "Failed to update the database with new schema: ")
		}
	} else {
		err = updateCurrentDB(schema)
		if err != nil {
			FatalError(err, "Failed to update the database with new schema: ")
		}
	}

}

// updateCurrentDB compares an existing SQLite table with the desired schema
// and migrates it to match. It adds missing columns when possible, and rebuilds
// the table when incompatible changes are detected
func updateCurrentDB(schema []UserColumn) error {
	exists, err := tableExists()
	if err != nil {
		return err
	}
	if !exists {
		return createTable(schema)
	}

	current, err := readCurrentSchema()
	if err != nil {
		return err
	}

	desiredMap := map[string]UserColumn{}
	for _, column := range schema {
		desiredMap[column.columnName] = column
	}

	currentMap := map[string]UserColumn{}
	for _, column := range current {
		currentMap[column.columnName] = column
	}

	needsRebuild := false

	// checks if theres an extra column in teh database (not mucho gracias)
	for name := range currentMap {
		if _, ok := desiredMap[name]; !ok {
			needsRebuild = true
			break
		}
	}

	// see if existing columns dont like changes in schema
	if !needsRebuild {
		for _, want := range schema {
			got, ok := currentMap[want.columnName]
			if !ok {
				continue
			}

			if !sameColumn(got, want) {
				needsRebuild = true
				break
			}
		}
	}

	if needsRebuild {
		return rebuildTable(schema, currentMap)
	}

	for _, want := range schema {
		if _, ok := currentMap[want.columnName]; ok {
			continue
		}

		if _, err := userDB.Exec("ALTER TABLE users ADD COLUMN " + buildColumnDef(want)); err != nil {
			return fmt.Errorf("add column %s: %w", want.columnName, err)
		}
	}

	return nil
}

func tableExists() (bool, error) {
	var n int
	err := userDB.QueryRow(
		`SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name=?`,
		"users",
	).Scan(&n)

	return n > 0, err
}

func createTable(schema []UserColumn) error {
	columnConstructors := make([]string, 0, len(schema))
	for _, column := range schema {
		columnConstructors = append(columnConstructors, buildColumnDef(column))
	}

	_, err := userDB.Exec(fmt.Sprintf("CREATE TABLE users (%s)", strings.Join(columnConstructors, ", ")))
	return err
}

// gets a user column array from the loaded database
func readCurrentSchema() ([]UserColumn, error) {
	rows, err := userDB.Query("PRAGMA table_info(users)")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []UserColumn
	for rows.Next() {
		var cid int
		var name, typee string // more like bladee
		var notNull, primaryKey int
		var defaultValue sql.NullString

		if err := rows.Scan(&cid, &name, &typee, &notNull, &defaultValue, &primaryKey); err != nil {
			return nil, err
		}

		cols = append(cols, UserColumn{
			columnName:   name,
			valueType:    strings.ToUpper(strings.TrimSpace(typee)),
			defaultValue: defaultValue,
			notNull:      notNull == 1,
			primaryKey:   primaryKey == 1,
			unique:       false, // filled below
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	uniqueColumns, err := readUniqueColumns()
	if err != nil {
		return nil, err
	}
	for i := range cols {
		if uniqueColumns[cols[i].columnName] {
			cols[i].unique = true
		}
	}

	return cols, nil
}

func readUniqueColumns() (map[string]bool, error) {
	result := map[string]bool{}

	idxRows, err := userDB.Query("PRAGMA index_list(users)")
	if err != nil {
		return nil, err
	}
	defer idxRows.Close()

	type idx struct {
		name   string
		unique int
	}
	var idxs []idx
	for idxRows.Next() {
		var seq int
		var name string
		var unique int
		var origin string
		var partial int
		if err := idxRows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			return nil, err
		}
		if unique == 1 {
			idxs = append(idxs, idx{name: name, unique: unique})
		}
	}
	if err := idxRows.Err(); err != nil {
		return nil, err
	}

	for _, ix := range idxs {
		infoRows, err := userDB.Query(fmt.Sprintf("PRAGMA index_info(%s)", friendlySqlStr(ix.name)))
		if err != nil {
			return nil, err
		}

		var columns []string
		for infoRows.Next() {
			var seqno, cid int
			var colName string
			if err := infoRows.Scan(&seqno, &cid, &colName); err != nil {
				infoRows.Close()
				return nil, err
			}
			columns = append(columns, colName)
		}
		infoRows.Close()

		if len(columns) == 1 {
			result[columns[0]] = true
		}
	}

	return result, nil
}

// compares two UserColumns against each other to if theyre the same
func sameColumn(oldColumn UserColumn, newColumn UserColumn) bool {
	if !strings.EqualFold(strings.TrimSpace(oldColumn.valueType), strings.TrimSpace(newColumn.valueType)) {
		return false
	}
	if oldColumn.notNull != newColumn.notNull {
		return false
	}
	if oldColumn.primaryKey != newColumn.primaryKey {
		return false
	}
	if oldColumn.unique != newColumn.unique {
		return false
	}

	gotDefault := ""
	if oldColumn.defaultValue.Valid {
		gotDefault = normalizeDefault(oldColumn.defaultValue.String)
	}
	wantDef := normalizeDefault(newColumn.defaultValue.String)
	return gotDefault == wantDef
}

func rebuildTable(schema []UserColumn, currentSchema map[string]UserColumn) error {
	tempName := "users__new"

	tx, err := userDB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	columnConstructors := make([]string, 0, len(schema))
	for _, c := range schema {
		columnConstructors = append(columnConstructors, buildColumnDef(c))
	}
	if _, err = tx.Exec(fmt.Sprintf("CREATE TABLE %s (%s)", friendlySqlStr(tempName), strings.Join(columnConstructors, ", "))); err != nil {
		return err
	}

	// copies over the existing columns
	var shared []string
	for _, column := range schema {
		if _, ok := currentSchema[column.columnName]; ok {
			shared = append(shared, friendlySqlStr(column.columnName))
		}
	}
	if len(shared) > 0 {
		columns := strings.Join(shared, ", ")
		if _, err = tx.Exec(fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM users", friendlySqlStr(tempName), columns, columns)); err != nil {
			return err
		}
	}

	if _, err = tx.Exec("DROP TABLE users"); err != nil {
		return err
	}
	if _, err = tx.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO users", friendlySqlStr(tempName))); err != nil {
		return err
	}

	for _, column := range schema {
		if column.unique && !column.primaryKey {
			idxName := fmt.Sprintf("idx_users_%s_unique", column.columnName)
			stmt := fmt.Sprintf(
				"CREATE UNIQUE INDEX IF NOT EXISTS %s ON users (%s)",
				friendlySqlStr(idxName), friendlySqlStr(column.columnName),
			)
			if _, err = tx.Exec(stmt); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func buildColumnDef(column UserColumn) string {
	parts := []string{friendlySqlStr(column.columnName), strings.ToUpper(strings.TrimSpace(column.valueType))}
	if column.primaryKey {
		parts = append(parts, "PRIMARY KEY")
	}
	if column.notNull {
		parts = append(parts, "NOT NULL")
	}
	if column.defaultValue.String != "" {
		parts = append(parts, "DEFAULT "+column.defaultValue.String)
	}
	if column.unique && !column.primaryKey {
		parts = append(parts, "UNIQUE")
	}
	return strings.Join(parts, " ")
}

// jarvis make sure my strings make SQL happy
func friendlySqlStr(input string) string {
	return `"` + strings.ReplaceAll(input, `"`, `""`) + `"`
}

func normalizeDefault(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "(")
	s = strings.TrimSuffix(s, ")")
	return strings.TrimSpace(s)
}

// creates a table from the schema into the users database
func populateNewDB(schema []UserColumn) error {
	columnConstructors := make([]string, 0, len(schema))
	for _, column := range schema {
		columnConstructors = append(columnConstructors, buildColumnDef(column))
	}
	_, err := userDB.Exec(fmt.Sprintf("CREATE TABLE users (%s)", strings.Join(columnConstructors, ", ")))
	return err
}

// Creates a new user
func NewUser(username string, uuid string) {
	badgeBytes, marshalError := json.Marshal(emptyBadges())

	if marshalError != nil {
		LogError(marshalError, "Problem marshalling empty badge JSON")
	}

	//The only reason most of these columns don't have default values is that sqlite doesn't let you alter column default values and I don't feel like deleting and remaking every column
	_, err := userDB.Exec("insert into users values(?,?,?,?,?,?,?, 0, 0, ?, 0, ?)", uuid, username, username, nil, string(badgeBytes), 0, DefaultPfpPath, "[]", "light")

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

	// If we ever get more than 1, something is horribly wrong.
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
		NewUser(username, uuid.New().String()) //Empty UUID for assignment later
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
func SetPfp(uuid string, pfp string) {
	_, execErr := userDB.Exec("update users set pfp = ? where uuid = ?", pfp, uuid)

	if execErr != nil {
		LogErrorf(execErr, "Problem executing sql query UPDATE users SET pfp = ? WHERE uuid = ? with args: %v, %v", pfp, uuid)
	}
}

// Gets the relative path of a given user's profile picture
func GetTheme(uuid string) string {
	var theme string
	response := userDB.QueryRow("select theme from users where uuid = ?", uuid)
	scanErr := response.Scan(&theme)
	if scanErr != nil {
		// LogError(scanErr, "Problem scanning response to sql query SELECT theme FROM users WHERE uuid = ? with arg: "+uuid)
		SetTheme(uuid, "light")
		return "light"
	}

	return theme
}

// Gets the relative path of a given user's profile picture
func SetTheme(uuid string, themeName string) {
	_, execErr := userDB.Exec("update users set theme = ? where uuid = ?", themeName, uuid)

	if execErr != nil {
		LogErrorf(execErr, "Problem executing sql query UPDATE users SET theme = ? WHERE uuid = ? with args: %v, %v", themeName, uuid)
	}
}
