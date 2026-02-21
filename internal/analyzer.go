package internal

// Utility for analyzing differences in cycle times

import (
	"math"

	"github.com/montanaflynn/stats"
)

// The allowable difference between cycle time averages
const kAllowableSeconds = 1.0

// Returns if the cycles passed in were within the configured acceptable range
// of similarity (time-based)
func CompareCycles(data [][]Cycle) bool {

	var averages []float64
	for _, entry := range data {
		averages = append(averages, entry[len(entry)-1].Time/float64(len(entry)))
	}

	return isNearSeconds(averages, kAllowableSeconds)
}

// Returns if the values within the passed in array are all within the passed in allowbable error
// of each other.
// If any error is encountered, it will return false.
func isNearSeconds(averages []float64, allowableErr float64) bool {

	max, maxErr := stats.Max(averages)
	if maxErr != nil {
		LogErrorf(maxErr, "Error finding maximum of %v", averages)
		return false
	}
	min, minErr := stats.Min(averages)
	if minErr != nil {
		LogErrorf(minErr, "Error finding minimum of %v", averages)
		return false
	}

	return math.Abs(max-min) <= allowableErr
}
