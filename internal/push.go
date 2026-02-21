package internal

// Utility for the nighly push to github of the databases

import (
	"os/exec"
	"strings"
)

// Makes a commit to the databases, and then pushes them to the upstream. If there is an upstream mismatch, the push will simply fail.
func CommitAndPushDBs() {
	commitCommand := exec.Command("git", "commit", "-am", "Daily database sync")
	pushCommand := exec.Command("git", "push")

	// Switch dir to the db path
	commitCommand.Dir = "./" + CachedConfigs.PathToDatabases
	pushCommand.Dir = "./" + CachedConfigs.PathToDatabases

	commit, commitErr := commitCommand.Output()
	LogMessage("Response to committing daily DB sync: " + string(commit))

	if commitErr != nil && !strings.Contains(commitErr.Error(), "exit status 1") {
		LogErrorf(commitErr, "Error Committing daily databases sync")
	} else {
		push, pushErr := pushCommand.Output()
		LogMessage("Response to pushing daily DB sync: " + string(push))

		if pushErr != nil && !strings.Contains(pushErr.Error(), "exit status 1") {
			LogErrorf(pushErr, "Error pushing daily databases sync")
		}
	}

}
