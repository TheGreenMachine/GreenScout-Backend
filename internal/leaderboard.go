package internal

// Utilities for interacting with leaderboards

import (
	"database/sql"
	"errors"
)

// A leaderboard modification request
type ModRequest struct {
	Name string       // The scouter username
	By   int          // How much to alter the score by
	Mod  Modification // How to alter the score
}

type Modification string

// Modification enum
const (
	Increase Modification = "Increase"
	Decrease Modification = "Decrease"
	Set      Modification = "Set"
)

// Modifies the score of a user by the passed in parameters.
func ModifyUserScore(name string, alter Modification, by int) {
	uuid, _ := GetUUID(name, true)

	switch alter {
	case Increase:
		var lifeScore int
		scoreRow := userDB.QueryRow("update users set score = score + ?, lifescore = lifescore + ? where uuid = ? returning lifescore", by, by, uuid)
		scanErr := scoreRow.Scan(&lifeScore)

		if scanErr != nil {
			LogErrorf(scanErr, "Problem scanning results of sql query UPDATE users SET score = score + %v, lifescore = lifescore + %v WHERE uuid = %v RETURNING score with args: ", by, by, uuid)
		}

		accolades := GetAccolades(uuid)

		if lifeScore >= 1 && !AccoladesHas(accolades, Rookie) {
			AddAccolade(uuid, Rookie, false)
		}

		if lifeScore >= 10 && !AccoladesHas(accolades, Novice) {
			AddAccolade(uuid, Novice, false)
		}

		if lifeScore >= 50 && !AccoladesHas(accolades, Scouter) {
			AddAccolade(uuid, Scouter, false)
		}

		if lifeScore >= 100 && !AccoladesHas(accolades, Pro) {
			AddAccolade(uuid, Pro, false)
		}

		if lifeScore >= 500 && !AccoladesHas(accolades, Enthusiast) {
			AddAccolade(uuid, Enthusiast, false)
			SetColor(uuid, Gold)
		}

	case Decrease:
		_, err := userDB.Exec("update users set score = score - ?, lifescore = lifescore - ? where uuid = ?", by, by, uuid)
		if err != nil {
			LogErrorf(err, "Problem executing sql query UPDATE users SET score = score - ?, lifescore = lifescore - ? WHERE uuid = ? with args: %v, %v, %v", by, by, uuid)
		}
	case Set: //Set will not affect life scores. Sorry!
		_, err := userDB.Exec("update users set score = ? where uuid = ?", by, uuid)
		if err != nil {
			LogErrorf(err, "Problem executing sql query UPDATE users SET score = ? WHERE uuid = ? with args: %v, %v", by, uuid)
		}
	}
	checkAndUpdateHighScore(uuid)
}

// Updates the high score if the current score is larger than it.
// TODO add storage & querying of the event the high score was achieved at. Learning experience for the next dev.
func checkAndUpdateHighScore(uuid string) {
	result := userDB.QueryRow("select score > highscore, score from users where uuid = ?", uuid)

	var larger bool
	var score int

	scanErr := result.Scan(&larger, &score)
	if scanErr != nil && !errors.Is(scanErr, sql.ErrNoRows) {
		LogErrorf(scanErr, "Problem scanning response %v", result)
	}

	if larger {

		_, execErr := userDB.Exec("update users set highscore = score where uuid = ?", uuid)
		if execErr != nil {
			LogErrorf(execErr, "Problem executing sql query UPDATE users SET highscore = score WHERE uuid = ? with arg: %v", uuid)
		}

		accolades := GetAccolades(uuid)

		if score >= 50 && !AccoladesHas(accolades, Locked) {
			AddAccolade(uuid, Locked, false)
		}

		if score >= 78 && !AccoladesHas(accolades, Deja) {
			AddAccolade(uuid, Deja, false)
		}

		if score >= 300 && !AccoladesHas(accolades, Eyes) {
			AddAccolade(uuid, Eyes, false)
			SetColor(uuid, Green)
		}
	}
}

// Returns the leaderboard ordered by the passed in score type
func GetLeaderboard(scoreType string) []UserInfo {
	var leaderboard []UserInfo

	// For some reason, the question mark doesn't work here.
	resultRows, queryErr := userDB.Query("select uuid, username, displayname, score, lifescore, highscore, color from users order by " + scoreType + " desc")

	if queryErr != nil {
		LogErrorf(queryErr, "Problem executing sql query SELECT uuid, username, displayname, score, lifescore, highscore, color FROM users ORDER BY %v DESC", scoreType)
	}

	for resultRows.Next() {
		var uuid string
		var username string
		var displayName string
		var score int
		var lifeScore int
		var highScore int
		var color int

		scanErr := resultRows.Scan(&uuid, &username, &displayName, &score, &lifeScore, &highScore, &color)

		if scanErr != nil {
			LogErrorf(scanErr, "Problem scanning response to sql query SELECT uuid, username, displayname, score, lifescore, highscore, color FROM users ORDER BY %v DESC", scoreType)
		}

		leaderboard = append(leaderboard, UserInfo{
			Username:    username,
			DisplayName: displayName,
			Badges:      GetBadges(uuid),
			Score:       score,
			LifeScore:   lifeScore,
			HighScore:   highScore,
			Color:       LBColor(color),
		})
	}

	return leaderboard
}

// Resets all current scores to 0
func ResetScores() {
	_, execErr := userDB.Exec("update users set score = 0") //No need to move anything around, as lifetime and high score are updated when score is added/set any other way
	if execErr != nil {
		LogErrorf(execErr, "Problem executing sql query UPDATE users SET score = 0")
	}
}
