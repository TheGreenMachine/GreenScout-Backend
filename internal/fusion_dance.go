package internal

import (
	"fmt"
	"math"

	"github.com/montanaflynn/stats"
)

// Utility for merging multiple MatchData instances into data to be written to the spreadsheet when multi-scouting

// Compliled data for an entire match from multiple scouters
type MultiMatch struct { // TODO: FIX FOR NEW
	TeamNumber    uint64             `json:"Team"`  // The team number
	Match         MatchInfo          `json:"Match"` // The match number
	Scouters      string             // The scouters who scouted this entry
	DriverStation DriverStationData  `json:"Driver Station"` // The driverstation of this entry
	CycleData     CompositeCycleData // The compiled cycle data from multiple scouters
	// Pickups       PickupLocations    // The compiled pickup locations from multiple scouters
	Auto   AutoDataV2 // The compiled auto data from multiple scouters
	Parked bool       // If any scouter recorded a park
	Notes  []string   // The compiled notes from multiple scouters
}

// Compiled scouting data from multiple scouters
type CompositeCycleData struct {
	NumCycles     int     // The computed number of cycles
	AvgCycleTime  float64 // The average cycle time
	AllCycles     []Cycle // All cycles raw
	HadMismatches bool    // If there were any mismatches
}

// Compiles Teamdata entries into one MultiMatch
func CompileMultiMatch(entries ...TeamDataV2) MultiMatch { // TODO: FIX FOR NEW
	var finalData MultiMatch

	// guy who uses c#: heh system.linq could this in 1/3 of the code

	teamNum, _ := compositeTeamNum(entries)

	finalData.TeamNumber = uint64(teamNum)

	finalData.Match = entries[0].Match

	finalData.Scouters = compositeScouters(entries)

	finalData.DriverStation = entries[0].DriverStation

	finalData.CycleData = compileCycles(entries)

	// finalData.Pickups = compilePickupPositions(entries)

	finalData.Auto = compileAutoData(entries)

	//TODO: DO MULTISCOUTING ENDGAME -Leon

	// finalData.Parked = compileParked(entries)

	finalData.Notes = compileNotes(entries, nil)

	return finalData
}

// Compiles the team number of all entries passed in. Always returns the first team number, as well as wether or not there were any mismatches
func compositeTeamNum(entries []TeamDataV2) (int, bool) {
	initial := entries[0].TeamNumber

	for i := 1; i < len(entries); i++ {
		if initial != entries[i].TeamNumber {
			return int(initial), true
		}
	}

	return int(initial), false
}

// Compiles the scouter names from all matches
func compositeScouters(entries []TeamDataV2) string {
	var finalScouter string
	for _, entry := range entries {
		finalScouter += fmt.Sprintf(", %s", entry.Scouter)
	}

	return finalScouter
}

// Compiles the cycle data from all matches into one CompositeCycleData
func compileCycles(entries []TeamDataV2) CompositeCycleData {
	var finalCycles CompositeCycleData
	var allNumCycles []int
	for _, entry := range entries {
		allNumCycles = append(allNumCycles, GetNumCycles(entry.Cycles))
	}

	for _, cycleNum := range allNumCycles {
		if cycleNum != allNumCycles[0] {
			finalCycles.HadMismatches = true
		}
	}

	cycleCompositeTime, hadMismatches := avgCycleTimes(entries)

	finalCycles.AvgCycleTime = cycleCompositeTime

	if hadMismatches {
		finalCycles.HadMismatches = true
	}

	var massiveBlockOfCycles []Cycle
	for _, entry := range entries {
		massiveBlockOfCycles = append(massiveBlockOfCycles, entry.Cycles...)
	}

	finalCycles.AllCycles = massiveBlockOfCycles

	return finalCycles
}

// Averages out the cycle times from all entries, returning this average as well as if there were any times that were outside
// of the configured acceptable range
func avgCycleTimes(entries []TeamDataV2) (float64, bool) {
	var sum float64
	var count int = 0

	var allCycles [][]Cycle

	for _, entry := range entries {
		allCycles = append(allCycles, entry.Cycles)
		entryAvg := GetAvgCycleTimeExclusive(entry.Cycles)
		if entryAvg != 0 {
			sum += entryAvg
			count++
		}
	}

	finalAvg := sum / float64(count)

	if math.IsNaN(finalAvg) {
		finalAvg = 0
	}
	return finalAvg, !CompareCycles(allCycles)
}

// Combines the pickup locations from all entries
// func compilePickupPositions(entries []TeamData) PickupLocations {
// 	var cGround bool = false
// 	var cSource bool = false
// 	var ground bool = false
// 	var source bool = false

// 	for _, entry := range entries {
// 		if entry.Pickups.AlgaeGround {
// 			ground = true
// 		}

// 		if entry.Pickups.AlgaeSource {
// 			source = true
// 		}

// 		if entry.Pickups.CoralGround {
// 			cGround = true
// 		}

// 		if entry.Pickups.CoralSource {
// 			cSource = true
// 		}
// 	}

// 	return PickupLocations{
// 		CoralGround: cGround,
// 		CoralSource: cSource,
// 		AlgaeGround: ground,
// 		AlgaeSource: source,
// 	}
// }

// Compiles autonomous data from all entries
func compileAutoData(entries []TeamDataV2) AutoDataV2 {
	// No need to mess with return values if err, as the NaNs do that well enough.

	var can bool = false
	var hang bool = false
	var won bool = false

	var left bool = false
	var right bool = false
	var mid bool = false
	var top bool = false
	var bump bool = false
	var trench bool = false
	var didntCross bool = false
	var hp bool = false
	var fuel bool = false

	var allScores []float64
	var allMisses []float64
	var allEjects []float64
	var allHumanAccuracy []float64
	var allRobotAccuracy []float64

	for _, entry := range entries {
		if entry.Auto.CanAuto {
			can = true
		}
		if entry.Auto.HangAuto {
			hang = true
		}
		if entry.Auto.WonAuto {
			won = true
		}

		// Field booleans: if any scouter marked it, keep it.
		if entry.Auto.Field.Left {
			left = true
		}
		if entry.Auto.Field.Right {
			right = true
		}
		if entry.Auto.Field.Mid {
			mid = true
		}
		if entry.Auto.Field.Top {
			top = true
		}
		if entry.Auto.Field.Bump {
			bump = true
		}
		if entry.Auto.Field.Trench {
			trench = true
		}
		if entry.Auto.Field.DidntCross {
			didntCross = true
		}
		if entry.Auto.Field.HP {
			hp = true
		}
		if entry.Auto.Field.Fuel {
			fuel = true
		}

		allScores = append(allScores, float64(entry.Auto.Scores))
		allMisses = append(allMisses, float64(entry.Auto.Misses))
		allEjects = append(allEjects, float64(entry.Auto.Ejects))

		allHumanAccuracy = append(allHumanAccuracy, float64(entry.Auto.Accuracy.HPAccuracy))
		allRobotAccuracy = append(allRobotAccuracy, float64(entry.Auto.Accuracy.RobotAccuracy))
	}

	scoresAvgd, scoresMeanErr := stats.Mean(allScores)
	if scoresMeanErr != nil {
		LogErrorf(scoresMeanErr, "Error finding mean of %v for all scores", allScores)
	}

	missesAvgd, missesMeanErr := stats.Mean(allMisses)
	if missesMeanErr != nil {
		LogErrorf(missesMeanErr, "Error finding mean of %v for all misses", allMisses)
	}

	ejectsAvgd, ejectsMeanErr := stats.Mean(allEjects)
	if ejectsMeanErr != nil {
		LogErrorf(ejectsMeanErr, "Error finding mean of %v for all ejects", allEjects)
	}

	hpAccAvgd, hpAccMeanErr := stats.Mean(allHumanAccuracy)
	if hpAccMeanErr != nil {
		LogErrorf(hpAccMeanErr, "Error finding mean of %v for all HP accuracy", allHumanAccuracy)
	}

	robotAccAvgd, robotAccMeanErr := stats.Mean(allRobotAccuracy)
	if robotAccMeanErr != nil {
		LogErrorf(robotAccMeanErr, "Error finding mean of %v for all robot accuracy", allRobotAccuracy)
	}

	return AutoDataV2{
		CanAuto:  can,
		HangAuto: hang,
		Scores:   int(scoresAvgd),
		Misses:   int(missesAvgd),
		Ejects:   int(ejectsAvgd),
		WonAuto:  won,

		Accuracy: AutoAccuracyV2{
			HPAccuracy:    int(hpAccAvgd),
			RobotAccuracy: int(robotAccAvgd),
		},
		Field: AutoFieldV2{
			Left:       left,
			Right:      right,
			Mid:        mid,
			Top:        top,
			Bump:       bump,
			Trench:     trench,
			DidntCross: didntCross,
			HP:         hp,
			Fuel:       fuel,
		},
	}
}

//TODO: ENDGAME compile MULTI -Leon

// Returns if any scouter recorded a park
// func compileParked(entries []TeamData) bool {
// 	for _, entry := range entries {
// 		if entry.Endgame.ParkStatus > 3 {
// 			return true
// 		}
// 	}
// 	return false
// }

// Combines the notes from all passed in scouters
func compileNotes(entries []TeamDataV2, mismatches []string) []string {
	var finalNotes []string
	for _, entry := range entries {
		combined := fmt.Sprintf("%s; %s; %s; %s; %s", entry.Notes.Auto, entry.Notes.Teleop, entry.Notes.Perf, entry.Notes.Events, entry.Notes.Comments)

		finalNotes = append(finalNotes, combined)
		finalNotes = append(finalNotes, mismatches...)
	}
	return finalNotes
}
