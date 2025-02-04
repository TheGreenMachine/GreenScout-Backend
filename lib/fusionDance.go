package lib

import (
	greenlogger "GreenScoutBackend/greenLogger"
	"errors"
	"fmt"
	"math"

	"github.com/montanaflynn/stats"
)

// Utility for merging multiple MatchData instances into data to be written to the spreadsheet when multi-scouting

// Compliled data for an entire match from multiple scouters
type MultiMatch struct {
	TeamNumber    uint64             `json:"Team"`  // The team number
	Match         MatchInfo          `json:"Match"` // The match number
	Scouters      string             // The scouters who scouted this entry
	DriverStation DriverStationData  `json:"Driver Station"` // The driverstation of this entry
	CycleData     CompositeCycleData // The compiled cycle data from multiple scouters
	Pickups       PickupLocations    // The compiled pickup locations from multiple scouters
	Auto          AutoData           // The compiled auto data from multiple scouters
	Climb         ClimbingData       // The compiled climbing data from multiple scouters
	Parked        bool               // If any scouter recorded a park
	TrapScore     int                // The compiled trap score from multiple scouters
	Notes         []string           // The compiled notes from multiple scouters
}

// Compiled scouting data from multiple scouters
type CompositeCycleData struct {
	NumCycles     int     // The computed number of cycles
	AvgCycleTime  float64 // The average cycle time
	AllCycles     []Cycle // All cycles raw
	HadMismatches bool    // If there were any mismatches
}

// Compiles Teamdata entries into one MultiMatch
func CompileMultiMatch(entries ...TeamData) MultiMatch {
	var finalData MultiMatch

	teamNum, _ := compositeTeamNum(entries)

	finalData.TeamNumber = uint64(teamNum)

	finalData.Match = entries[0].Match

	finalData.Scouters = compositeScouters(entries)

	finalData.DriverStation = entries[0].DriverStation

	finalData.CycleData = compileCycles(entries)

	finalData.Pickups = compilePickupPositions(entries)

	finalData.Auto = compileAutoData(entries)

	finalData.Climb = compileClimbData(entries)

	finalData.Parked = compileParked(entries)

	finalData.TrapScore = compileTrapScore(entries)

	finalData.Notes = compileNotes(entries, nil)

	return finalData
}

// Compiles the team number of all entries passed in. Always returns the first team number, as well as wether or not there were any mismatches
func compositeTeamNum(entries []TeamData) (int, bool) {
	initial := entries[0].TeamNumber

	for i := 1; i < len(entries); i++ {
		if initial != entries[i].TeamNumber {
			return int(initial), true
		}
	}

	return int(initial), false
}

// Compiles the scouter names from all matches
func compositeScouters(entries []TeamData) string {
	var finalScouter string
	for _, entry := range entries {
		finalScouter += fmt.Sprintf(", %s", entry.Scouter)
	}

	return finalScouter
}

// Compiles the cycle data from all matches into one CompositeCycleData
func compileCycles(entries []TeamData) CompositeCycleData {
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
func avgCycleTimes(entries []TeamData) (float64, bool) {
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
func compilePickupPositions(entries []TeamData) PickupLocations {
	var cGround bool = false
	var cSource bool = false
	var ground bool = false
	var source bool = false

	for _, entry := range entries {
		if entry.Pickups.AlgaeGround {
			ground = true
		}

		if entry.Pickups.AlgaeSource {
			source = true
		}

		if entry.Pickups.CoralGround {
			cGround = true
		}

		if entry.Pickups.CoralSource {
			cSource = true
		}
	}

	return PickupLocations{
		CoralGround: cGround,
		CoralSource: cSource,
		AlgaeGround: ground,
		AlgaeSource: source,
	}
}

// Compiles autonomous data from all entries
func compileAutoData(entries []TeamData) AutoData {
	// No need to mess with return values if err, as the NaNs do that well enough.

	var can bool = false
	var allScores []float64
	var allMisses []float64
	var allEjects []float64

	for _, entry := range entries {
		if entry.Auto.Can {
			can = true
		}

		allScores = append(allScores, float64(entry.Auto.Scores))
		allMisses = append(allMisses, float64(entry.Auto.Misses))
		allEjects = append(allEjects, float64(entry.Auto.Ejects))
	}

	scoresAvgd, scoresMeanErr := stats.Mean(allScores)
	if scoresMeanErr != nil {
		greenlogger.LogErrorf(scoresMeanErr, "Error finding mean of %v for all scores", allScores)
	}

	missesAvgd, missesMeanErr := stats.Mean(allMisses)
	if missesMeanErr != nil {
		greenlogger.LogErrorf(missesMeanErr, "Error finding mean of %v for all misses", allMisses)
	}

	ejectsAvgd, ejectsMeanErr := stats.Mean(allEjects)
	if ejectsMeanErr != nil {
		greenlogger.LogErrorf(ejectsMeanErr, "Error finding mean of %v for all ejects", allEjects)
	}

	return AutoData{
		Can:    can,
		Scores: int(scoresAvgd),
		Misses: int(missesAvgd),
		Ejects: int(ejectsAvgd),
	}
}

// Compiles climbing data from all entries
func compileClimbData(entries []TeamData) ClimbingData {
	var success bool = false
	var times []float64

	for _, entry := range entries {
		if entry.Climb.Succeeded {
			success = true
		}

		if entry.Climb.Time > 0 {
			times = append(times, entry.Climb.Time)
		}
	}
	timeAvgd, err := stats.Mean(times)

	if err != nil {
		if errors.Is(err, stats.EmptyInput) {
			timeAvgd = 0
		} else {
			greenlogger.LogErrorf(err, "Error finding mean of %v for climb times", times)
		}
	}

	return ClimbingData{
		Succeeded: success,
		Time:      timeAvgd,
	}
}

// Returns if any scouter recorded a park
func compileParked(entries []TeamData) bool {
	for _, entry := range entries {
		if entry.EndgameData.parkStatus > 3 {
			return true
		}
	}
	return false
}

// Averages out the trap scores from all scouters
func compileTrapScore(entries []TeamData) int {
	var trapScores []float64
	for _, entry := range entries {
		trapScores = append(trapScores, float64(entry.Trap.Score))
	}

	trapAvgd, err := stats.Mean(trapScores)
	if err != nil {
		greenlogger.LogErrorf(err, "Error finding mean of %v for trap scores", trapScores)
	}

	return int(math.Round(trapAvgd))
}

// Combines the notes from all passed in scouters
func compileNotes(entries []TeamData, mismatches []string) []string {
	var finalNotes []string
	for _, entry := range entries {
		finalNotes = append(finalNotes, entry.Notes)
		finalNotes = append(finalNotes, mismatches...)
	}
	return finalNotes
}
