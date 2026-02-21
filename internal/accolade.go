package internal

// Utilities for handling accolades/achievments/badges

import (
	"encoding/json"
	"slices"
)

// Accolade name enum. Not entirely accurate.
const (
	Rookie      Accolade = "Scouting Rookie"     // Scouted 1
	Novice      Accolade = "Scouting Novice"     // Scouted 10
	Scouter     Accolade = "Scouter"             // Scouted 50
	Pro         Accolade = "Scouting Pro"        // Scouted 100
	Enthusiast  Accolade = "Scouting Enthusiast" // Scouted 500
	Locked      Accolade = "Locked In"           // High 50
	Deja        Accolade = "Déjà vu"             // High 78
	Eyes        Accolade = "👀"                   // YES THIS IS DUMB I KNOW; High 300
	Strategizer Accolade = "Strategizer"
	Foreign     Accolade = "Foreign Fracas"
	Detective   Accolade = "Detective"
	Debugger    Accolade = "Debugger"
	mnmi224     Accolade = "Mvp_2024mnmi2" //This one is wrong but i forgot the right thing
	Dev         Accolade = "App Dev"
	FDev        Accolade = "Frontend Dev"
	BDev        Accolade = "Backend Dev"
	StratLead   Accolade = "Strategy Lead"
	Lead        Accolade = "Leadership"
	Captain     Accolade = "Captain"
	AsstCaptain Accolade = "Assistant Captain"
	CSPLead     Accolade = "CSP Lead"
	MechLead    Accolade = "Mechanical Lead"
	Mentor      Accolade = "Mentor"
	AppMentor   Accolade = "App Mentor"
	Admin       Accolade = "Admin"
	Super       Accolade = "Super Admin"
	Test        Accolade = "Test"
	Driveteam   Accolade = "Driveteam"
	HOF         Accolade = "HOF"
	Bug         Accolade = "Bug Finder"
	Early       Accolade = "Early"
	Router      Accolade = "Router Dungeon Survivor"
)

// All accolades in array form for easy comparison
var AllAccolades = []Accolade{
	Rookie,
	Novice,
	Scouter,
	Pro,
	Enthusiast,
	Locked,
	Deja,
	Eyes,
	Strategizer,
	Foreign,
	Detective,
	Debugger,
	mnmi224,
	Dev,
	FDev,
	BDev,
	StratLead,
	Lead,
	Captain,
	AsstCaptain,
	CSPLead,
	MechLead,
	Mentor,
	AppMentor,
	Admin,
	Super,
	Test,
	Driveteam,
	HOF,
	Bug,
	Early,
	Router,
}

// All achievments that are frontend-assigned
var frontendAchievements = []Accolade{
	Strategizer,
	Foreign,
	Detective,
	Debugger,
}

// Struct for requests for adding accolades from the frontend
type FrontendAdds struct {
	UUID         string     `json:"UUID"`
	Achievements []Accolade `json:"Achievements"`
}

// Consumes and processes Frontend Additions
func ConsumeFrontendAdditions(adds FrontendAdds, isAdmin bool) {
	for _, add := range adds.Achievements {
		if isAdmin && slices.Contains(AllAccolades, add) { // Theoretically, admins can add any achievement. Never implemented this on the frontend, so it's CLI-only
			AddAccolade(adds.UUID, add, false)
		} else if slices.Contains(frontendAchievements, add) {
			AddAccolade(adds.UUID, add, true)
		}
	}
}

// Gets all accolades for a given user
func GetAccolades(uuid string) []AccoladeData {
	var Accolades []AccoladeData
	var AccoladesMarshalled string
	response := userDB.QueryRow("select accolades from users where uuid = ?", uuid)
	scanErr := response.Scan(&AccoladesMarshalled)
	if scanErr != nil {
		LogError(scanErr, "Problem scanning results of sql query SELECT accolades FROM users WHERE uuid = ? with arg: "+uuid)
	}
	// I am aware of how awful converting []byte -> string -> []byte is but i've had problems storing byte arrays with sqlite.
	unmarshalErr := json.Unmarshal([]byte(AccoladesMarshalled), &Accolades)
	if unmarshalErr != nil {
		LogErrorf(unmarshalErr, "Problem unmarshalling %v", AccoladesMarshalled)
	}

	return Accolades
}

// Gets the acoolade names from an array of accolade data
func ExtractNames(accolades []AccoladeData) []Accolade {
	var names []Accolade
	for _, accolade := range accolades {
		names = append(names, accolade.Accolade)
	}
	return names
}

// Checks if an array of accolade data has a given accolade
func AccoladesHas(accolades []AccoladeData, accoladeToCheck Accolade) bool {
	for _, accolade := range accolades {
		if accolade.Accolade == accoladeToCheck {
			return true
		}
	}
	return false
}
