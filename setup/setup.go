package setup

// Handles server setup upon bootup

import (
	"GreenScoutBackend/constants"
	filemanager "GreenScoutBackend/fileManager"
	greenlogger "GreenScoutBackend/greenLogger"
	"GreenScoutBackend/lib"
	"GreenScoutBackend/rsaUtil"
	"GreenScoutBackend/schedule"
	"GreenScoutBackend/sheet"
	"GreenScoutBackend/userDB"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	yaml "sigs.k8s.io/yaml/goyaml.v2"
)

// I'm really sorry for how I named these functions. good luck.

// Runs through the entire setup routine
func TotalSetup(publicHosting bool) {
	// Config retrieval
	greenlogger.LogMessage("Retreiving configs...")
	configs := retrieveGeneralConfigs()
	greenlogger.LogMessagef("General configs retrieved: %v", configs)

	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Set frontend domain to be used for CORS
	if configs.FrontendDomain == "" {
		if !publicHosting {
			configs.FrontendDomain = constants.DefaultFrontendDomain
		} else {
			panic("Please configure a FrontendDomain when running in production!")
		}
	}

	// Initialize runtime directory and file paths
	if configs.RuntimeDirectory == "" {
		configs.RuntimeDirectory = filepath.Join(workingDir, constants.DefaultRuntimeDirectory)
	}
	if configs.JsonDirectory == "" {
		configs.JsonDirectory = filepath.Join(configs.RuntimeDirectory, constants.DefaultJsonDirectory)
	}
	if configs.TeamListsDirectory == "" {
		configs.TeamListsDirectory = filepath.Join(configs.RuntimeDirectory, constants.DefaultTeamsDirectory)
	}
	if configs.PfpDirectory == "" {
		configs.PfpDirectory = filepath.Join(configs.RuntimeDirectory, constants.DefaultPfpDirectory)
	}
	if configs.GalleryDirectory == "" {
		configs.GalleryDirectory = filepath.Join(configs.RuntimeDirectory, constants.DefaultGalleryDirectory)
	}
	if configs.CertsDirectory == "" {
		configs.CertsDirectory = filepath.Join(configs.RuntimeDirectory, constants.DefaultCertsDirectory)
	}

	constants.RSAPubKeyPath = filepath.Join(configs.RuntimeDirectory, "login-key.pub.pem")
	constants.RSAPrivateKeyPath = filepath.Join(configs.RuntimeDirectory, "login-key.pem")
	constants.DefaultPfpPath = filepath.Join(configs.PfpDirectory, constants.DefaultPfp)

	constants.JsonInDirectory = filepath.Join(configs.JsonDirectory, "In")
	constants.JsonWrittenDirectory = filepath.Join(configs.JsonDirectory, "Written")
	constants.JsonMangledDirectory = filepath.Join(configs.JsonDirectory, "Mangled")
	constants.JsonArchiveDirectory = filepath.Join(configs.JsonDirectory, "Archive")
	constants.JsonErroredDirectory = filepath.Join(configs.JsonDirectory, "Errored")
	constants.JsonDiscardedDirectory = filepath.Join(configs.JsonDirectory, "Discarded")
	constants.JsonPitWrittenDirectory = filepath.Join(configs.JsonDirectory, "PitWritten")

	// Essential Databases
	configs.PathToDatabases = filepath.Join(configs.RuntimeDirectory, constants.DefaultDbDirectory) //This is the only one i'm not having the user enter mainly because git cloning is uniform
	ensureDatabasesExist(configs)
	greenlogger.LogMessage("Essential databases verified...")

	// Sheets API
	greenlogger.LogMessage("Ensuring sheets API...")
	//ensureSheetsAPI(configs)
	greenlogger.LogMessage("Sheets API confirmed set-up")

	// Sqlite
	greenlogger.LogMessage("Ensuring sqlite3 driver...")
	configs.SqliteDriver = ensureSqliteDriver()
	greenlogger.LogMessagef("Sqlite driver validated: %v", configs.SqliteDriver)

	// Inputted JSON dirs
	greenlogger.LogMessage("Ensuring InputtedJSON...")
	ensureInputtedJSON()
	greenlogger.LogMessage("InputtedJSON folders confirmed to exist")

	// RSA key generation
	greenlogger.LogMessage("Ensuring RSA keys...")
	ensureRSAKey()
	greenlogger.LogMessage("RSA keys confirmed to exist")

	// Scout.db
	greenlogger.LogMessage("Ensuring scouting schedule database...")
	ensureScoutDB(configs)
	greenlogger.LogMessage("Schedule database confirmed to exist")

	// TBA API package
	greenlogger.LogMessage("Ensuring TBA API python package...")
	// downloadAPIPackage()
	greenlogger.LogMessage("API package present")

	// Network
	if publicHosting {
		// IP
		greenlogger.LogMessage("Ensuring ip in configs...")
		configs.IP = recursivelyEnsureIP(configs.IP)
		greenlogger.LogMessagef("IP %v confirmed ipv4", configs.IP)

		// Domain
		greenlogger.LogMessage("Ensuring domain name maps to IP...")
		configs.DomainName = recursivelyEnsureFunctionalDomain(&configs, configs.DomainName)
		greenlogger.LogMessagef("Domain %v confirmed to match IP %v", configs.DomainName, configs.IP)
	} else {
		// Allows stuff to go though localhost
		greenlogger.LogMessage("TEST MODE: Skipping ip and domain name ensuring...")
	}

	// Python
	greenlogger.LogMessage("Ensuring python driver...")
	// configs.PythonDriver = ensurePythonDriver(configs.PythonDriver)
	greenlogger.LogMessagef("Python driver validated: %v", configs.PythonDriver)

	// TBA API key
	greenlogger.LogMessage("Ensuring TBA API key...")
	// configs.TBAKey = ensureTBAKey(configs)
	greenlogger.LogMessagef("TBA key validated: %v", configs.TBAKey)

	// Event key
	greenlogger.LogMessage("Ensuring Event key...")
	configs.EventKey, configs.EventKeyName = ensureEventKey(configs)
	greenlogger.LogMessagef("Event key validated: %v", configs.EventKey)

	// Events
	greenlogger.LogMessage("Writing all events to file...")
	lib.WriteEventsToFile(configs)
	greenlogger.LogMessage("All events written to file")

	// More event config
	if !constants.CustomEventKey {
		/// TBA Event

		// Schedule
		greenlogger.LogMessage("Writing event schedule to file...")
		lib.WriteScheduleToFile(configs)
		greenlogger.LogMessage("Event schedule written to file")

		// Teamlist
		lib.WriteTeamsToFile(configs)
		greenlogger.LogMessagef("Teams at %v written to file", configs.EventKey)
	} else {
		/// Custom event
		configs.CustomEventConfigs = configCustomEvent(configs)
		if configs.CustomEventConfigs.PitScouting {
			// Teamlist
			if !lib.CheckForTeamLists(configs.EventKey) {
				greenlogger.FatalLogMessage(fmt.Sprintf("Please ensure that a Team List exists in %s for your event, as you plan to pit scout.", configs.TeamListsDirectory))
			}
		}
	}

	// Spreadsheet ID
	// configs.SpreadSheetID = recursivelyEnsureSpreadsheetID(configs.SpreadSheetID)
	greenlogger.LogMessagef("Spreadsheet ID %v verified...", configs.SpreadSheetID)

	// Logging
	if !configs.LogConfigs.Configured {
		configs.LogConfigs.Configured = true
		configs.LogConfigs.Logging = true
		configs.LogConfigs.LoggingHttp = true
	} else if !configs.LogConfigs.Logging {
		greenlogger.ShutdownLogFile()
	}

	/// writing

	configFile, openErr := filemanager.OpenWithPermissions(constants.ConfigFilePath)
	if openErr != nil {
		greenlogger.LogErrorf(openErr, "Problem creating %v", constants.ConfigFilePath)
	}

	defer configFile.Close()

	// Write back to yaml
	encodeErr := yaml.NewEncoder(configFile).Encode(&configs)

	if encodeErr != nil {
		greenlogger.LogErrorf(encodeErr, "Problem encoding %v", configs)
	}

	// Write to memory
	constants.CachedConfigs = configs

	greenlogger.LogMessagef("Setup finished! If you need to alter configurations any further, please check %v", constants.ConfigFilePath)
}

// Gets the general configs from yaml and returns a GeneralConfigs object containing them
func retrieveGeneralConfigs() constants.GeneralConfigs {
	var genConfigs constants.GeneralConfigs

	configFile, openErr := os.Open(constants.ConfigFilePath)
	if openErr != nil && !errors.Is(openErr, os.ErrNotExist) {
		greenlogger.LogErrorf(openErr, "Problem opening %v", constants.ConfigFilePath)
	}
	defer configFile.Close()

	dataAsByte, readErr := io.ReadAll(configFile)

	if readErr != nil && configFile != nil {
		greenlogger.LogErrorf(readErr, "Problem reading %v", configFile)
	}

	unmarshalErr := yaml.Unmarshal(dataAsByte, &genConfigs)
	if unmarshalErr != nil {
		greenlogger.LogErrorf(unmarshalErr, "Problem unmarshalling %v", dataAsByte)
	}
	return genConfigs
}

// Runs the python ensurance routine and returns the driver eventually
func ensurePythonDriver(existingDriver string) string {
	if validatePythonDriver(existingDriver) {
		return existingDriver
	}

	return recursivePythonValidation(true)
}

// Validates for the python driver. If for some reason the entered in driver doesn't validate, it will recurse and not return until it has a valid one
func recursivePythonValidation(firstRun bool) string {
	if firstRun {
		greenlogger.LogMessage("Enter the python driver installed on this machine (what you type to run a .py file from the command line): ")
	}

	var driver string
	_, scanErr := fmt.Scanln(&driver)

	if scanErr != nil {
		greenlogger.LogError(scanErr, "Problem scanning python driver input")
	}

	if validatePythonDriver(driver) {
		return driver
	} else {
		greenlogger.LogMessagef("Sorry, %v doesn't appear to be a valid python driver. Please ensure you didn't make a typo!", driver)
		return recursivePythonValidation(false)
	}
}

// Checks if the python driver is valid by checking for its version
func validatePythonDriver(driver string) bool {
	runnable := exec.Command(driver, "--version")

	out, execErr := runnable.Output()
	if execErr != nil && !strings.Contains(execErr.Error(), "no command") {
		greenlogger.LogErrorf(execErr, "Problem executing %v %v", driver, "--version")
	}

	return len(out) > 0 && strings.Contains(string(out), "Python")
}

// Checks if the sqlite3 driver exists on the machine and returns sqlite3. If not, it will fatal the program.
func ensureSqliteDriver() string {
	if !validateSqliteDriver() {
		greenlogger.FatalLogMessage("Invalid sqlite3 driver! Please ensure it's in your path and accessable to this program. \n If you don't have sqlite, please download it at https://www.sqlite.org/")
	}

	return "sqlite3"
}

// Uses regex validation and a call to sqlite3 -version to ensure it's installed on thsis machine
func validateSqliteDriver() bool {
	// Define the pattern to match 3.{someNumber}.{someNumber}
	pattern := `3\.\d+\.\d+`

	// This is so dumb why can't it just have sqlite in its name like every other -version arg
	re := regexp.MustCompile(pattern)

	runnable := exec.Command("sqlite3", "-version")

	out, execErr := runnable.Output()
	if execErr != nil {
		greenlogger.LogErrorf(execErr, "Problem executing %v %v", "sqlite3", "--version")
	}

	return re.FindString(string(out)) != ""
}

// Runs getStatus.py with the entered in TBA key, returning if it was successful.
// This is unreliable because TBA is very weird at times. It will sometimes let an incorrect api key authenticate, so please ensure you've got the right one.
func validateTBAKey(configs constants.GeneralConfigs, key string) bool {
	if key == "" {
		return false
	}

	runnable := exec.Command(configs.PythonDriver, "getStatus.py", key)

	out, execErr := runnable.Output()

	if execErr != nil {
		greenlogger.LogErrorf(execErr, "Problem executing %v %v %v", configs.PythonDriver, "getStatus.py", key)
	}

	return string(out) != "ERR"
}

// Runs the TBA-key ensurance routine, eventually returning the valid key
func ensureTBAKey(configs constants.GeneralConfigs) string {
	if validateTBAKey(configs, configs.TBAKey) {
		return configs.TBAKey
	}

	return recursiveTBAKeyValidation(&configs, true)
}

// Validates for the TBA API key. If the key is invalid, it will recurse and not return until it has a valid one
func recursiveTBAKeyValidation(configs *constants.GeneralConfigs, firstRun bool) string {
	if firstRun {
		greenlogger.LogMessage("Enter your Blue Alliance API Key: ")
	}

	var key string
	_, scanErr := fmt.Scanln(&key)

	if scanErr != nil {
		greenlogger.LogError(scanErr, "Problem scanning TBA key input")
	}

	if validateTBAKey(*configs, key) {
		return key
	} else {
		greenlogger.LogMessagef("Sorry, %v doesn't appear to be a valid TBA Key. ", key)
		return recursiveTBAKeyValidation(configs, false)
	}
}

// Validates for the TBA event key. If it is a custom event key, it will simply return that. If not, it will run getEvent.py and return its result.
func validateEventKey(configs constants.GeneralConfigs, key string) string {
	if len(key) != 0 {
		if string(key[0]) == "c" { // Check for custom event
			constants.CustomEventKey = true
			return configs.EventKeyName
		}
	}

	runnable := exec.Command(configs.PythonDriver, "getEvent.py", configs.TBAKey, key)

	out, execErr := runnable.Output()

	if execErr != nil {
		greenlogger.LogErrorf(execErr, "Problem executing %v %v %v %v", configs.PythonDriver, "getEvent.py", configs.TBAKey, key)
	}

	return string(out)
}

// Runs the event key ensurance routine. It will eventually return a valid TBA event key or a custom key.
func ensureEventKey(configs constants.GeneralConfigs) (string, string) {
	response := validateEventKey(configs, configs.EventKey)
	if !strings.Contains(response, "ERR") {
		configs.EventKeyName = strings.ReplaceAll(strings.Trim(response, "\n"), "'", "")

		return configs.EventKey, configs.EventKeyName
	}

	return recursiveEventKeyValidation(&configs, true)
}

// Recursively validates for the TBA API key. If the key is invalid, it will recurse and not return until it has a valid one
func recursiveEventKeyValidation(configs *constants.GeneralConfigs, firstRun bool) (string, string) {
	if firstRun {
		greenlogger.LogMessage("Please enter the Blue alliance Event Key to be used (ex: 2024mnst); For non-TBA events, please start your fake key with 'c' (ex: c2024gtch)")
	}

	var key string
	_, scanErr := fmt.Scanln(&key)

	if scanErr != nil {
		greenlogger.LogError(scanErr, "Problem scanning TBA key input")
	}

	if name := validateEventKey(*configs, key); !strings.Contains(name, "ERR") {
		moveOldJson(key)
		return key, strings.ReplaceAll(strings.Trim(name, "\n"), "'", "")
	} else {
		greenlogger.LogMessagef("Sorry, %v doesn't appear to be a valid Event Key. ", key)
		return recursiveEventKeyValidation(configs, false)
	}
}

// Handles setting the event key. If the passed in key is valid, it will change the cached configs, the file-encoded configs, and trigger
// writing to schedule.json, TeamLists, storing teams, and resetting user scores.
func SetEventKey(key string) bool {
	file, openErr := filemanager.OpenWithPermissions(constants.ConfigFilePath)
	if openErr != nil {
		greenlogger.LogErrorf(openErr, "Problem creating %v", constants.ConfigFilePath)
	}
	defer file.Close()

	if name := validateEventKey(constants.CachedConfigs, key); !strings.Contains(name, "ERR") {
		constants.CachedConfigs.EventKey = key
		constants.CachedConfigs.EventKeyName = strings.Trim(strings.ReplaceAll(name, "'", ""), "\n")

		encodeErr := yaml.NewEncoder(file).Encode(&constants.CachedConfigs)

		if encodeErr != nil {
			greenlogger.LogErrorf(encodeErr, "Problem encoding %v to %v", constants.CachedConfigs, constants.ConfigFilePath)
		}

		lib.WriteScheduleToFile(constants.CachedConfigs)
		lib.WriteTeamsToFile(constants.CachedConfigs)
		lib.StoreTeams()

		userDB.ResetScores()

		greenlogger.LogMessagef("Successfully changed Event Key to %v", key)

		return true
	}

	return false
}

// Makes all the directories of InputtedJson
func ensureInputtedJSON() {
	greenlogger.HandleMkdirAll(constants.JsonInDirectory)
	greenlogger.HandleMkdirAll(constants.JsonMangledDirectory)
	greenlogger.HandleMkdirAll(constants.JsonWrittenDirectory)
	greenlogger.HandleMkdirAll(constants.JsonArchiveDirectory)
	greenlogger.HandleMkdirAll(constants.JsonErroredDirectory)
	greenlogger.HandleMkdirAll(constants.JsonDiscardedDirectory)
	greenlogger.HandleMkdirAll(constants.JsonPitWrittenDirectory)
}

// Moves all JSON files from Written to Archive upon changes to the event key
func moveOldJson(newKey string) {
	allJson, readErr := os.ReadDir(constants.JsonWrittenDirectory)

	if readErr != nil {
		greenlogger.LogErrorf(readErr, "Problem reading %v", constants.JsonWrittenDirectory)
	}

	for _, file := range allJson {
		if !strings.Contains(file.Name(), newKey) { // If they aren't from this event
			newPath := filepath.Join(constants.JsonArchiveDirectory, strings.Split(file.Name(), "_")[0])
			greenlogger.HandleMkdirAll(newPath) // Archive folder

			oldStr := filepath.Join(constants.JsonInDirectory, file.Name())
			oldLoc, openErr := os.Open(oldStr)
			if openErr != nil {
				greenlogger.LogErrorf(openErr, "Problem opening %v", oldStr)
			}

			newLoc, openErr := filemanager.OpenWithPermissions(filepath.Join(newPath, file.Name()))

			if openErr != nil {
				greenlogger.LogErrorf(openErr, "Problem creating %v", filepath.Join(newPath, file.Name()))
			}

			defer newLoc.Close()

			// Copy old -> new
			_, copyErr := io.Copy(newLoc, oldLoc)

			if copyErr != nil {
				greenlogger.LogErrorf(copyErr, "Problem copying %v to %v", oldStr, filepath.Join(newPath, file.Name()))
			}

			oldLoc.Close()

			// Delete old
			removeErr := os.Remove(oldStr)

			if removeErr != nil {
				greenlogger.LogErrorf(removeErr, "Problem removing %v", oldStr)
			}
		}
	}
}

// Ensures the existence of the RSA keys used for asymetrically logging in. If they don't exist, it will generate them.
func ensureRSAKey() {
	if file, err := os.Open(constants.RSAPubKeyPath); errors.Is(err, os.ErrNotExist) {
		generateRSAPair()
		closeErr := file.Close()
		if closeErr != nil {
			greenlogger.LogErrorf(closeErr, "Problem closing %v", constants.RSAPubKeyPath)
		}
	} else if file, err := os.Open(constants.RSAPrivateKeyPath); errors.Is(err, os.ErrNotExist) {
		generateRSAPair()
		closeErr := file.Close()
		if closeErr != nil {
			greenlogger.LogErrorf(closeErr, "Problem closing %v", constants.RSAPrivateKeyPath)
		}
	}

	// Test if it can encode-decode successfully
	if rsaUtil.DecryptPassword(rsaUtil.EncodeWithPublicKey("test")) != "test" {
		greenlogger.FatalLogMessage("RSA keys mismatched! Look into this!")
	}

}

// Generates a pair of RSA keys for use in logging in.
func generateRSAPair() {
	bitSize := 4096

	// Generate RSA key.
	key, keyGenErr := rsa.GenerateKey(rand.Reader, bitSize)
	if keyGenErr != nil {
		greenlogger.FatalLogMessage(keyGenErr.Error())
	}

	// Extract public component.
	pub := key.Public()

	// Encode private key to PKCS#1 ASN.1 PEM.
	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	// Encode public key to PKCS#1 ASN.1 PEM.
	pubPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pub.(*rsa.PublicKey)),
		},
	)

	if err := filemanager.WriteFileWithPermissions(constants.RSAPrivateKeyPath, keyPEM); err != nil {
		greenlogger.FatalLogMessage(err.Error())
	}

	// Write public key to file.
	if err := filemanager.WriteFileWithPermissions(constants.RSAPubKeyPath, pubPEM); err != nil {
		greenlogger.FatalLogMessage(err.Error())
	}
}

// Ensures scout.db exists. If not, creates itt.
func ensureScoutDB(configs constants.GeneralConfigs) {

	_, err := os.Stat(filepath.Join(configs.RuntimeDirectory, "scout.db"))
	if err != nil && os.IsNotExist(err) && filemanager.IsSudo() {
		greenlogger.FatalLogMessage("scout.db must still be created, please run 'go run main.go setup' without sudo so you can alter its contents in the future.")
	}

	dbRef, openErr := sql.Open(configs.SqliteDriver, filepath.Join(configs.RuntimeDirectory, "scout.db"))

	if openErr != nil {
		greenlogger.FatalLogMessage(openErr.Error())
	}

	var response any
	// Counts up for every entry in individuals
	scanErr := dbRef.QueryRow("select count(1) from individuals").Scan(&response)
	if scanErr != nil {
		greenlogger.LogErrorf(scanErr, "Problem scanning SQL query result from %v", "select count(1) from individuals")
	}

	if response == nil { // If it wasn't able to count up, that means the table doesn't exist, so create it.
		_, execErr := dbRef.Exec("CREATE TABLE individuals(uuid string not null primary key, username string, schedule string)")

		if execErr != nil {
			greenlogger.FatalLogMessage("Problem creating scouting schedule database")
		}
	}

	closeErr := dbRef.Close()
	if closeErr != nil {
		greenlogger.LogError(closeErr, "Problem closing scouting schedule database")
	}
}

// Checks for credentials.json, required for the sheets API. If it doesn't exist, it will exit the program.
func ensureSheetsAPI(configs constants.GeneralConfigs) {
	creds, err := os.ReadFile(filepath.Join("conf", "credentials.json"))
	if err != nil {
		greenlogger.LogMessage("It appears there isn't a credentials.json file. Please follow the 'set up your environment' steps here: https://developers.google.com/sheets/api/quickstart/go#set_up_your_environment")
		greenlogger.LogMessage("Remember to publish your Google Cloud project before you create your tokens so that they don't expire after a few days!")
		greenlogger.FatalError(err, "Unable to read credentials file")
	}

	constants.SheetsTokenFile = filepath.Join(configs.RuntimeDirectory, "token.json")
	sheet.SetupSheetsAPI(creds)
}

// Checks if the passed in domain is valid and matches the IP in the passed in GeneralConfigs. If not, it will recurse until it finally returns a valid one.
func recursivelyEnsureFunctionalDomain(configs *constants.GeneralConfigs, domain string) string {
	res, lookupErr := net.LookupIP(domain)

	if lookupErr != nil && domain != "" {
		greenlogger.FatalLogMessage("Unable to look up domain " + domain)
	}

	// Check for the IP mapping to the domain
	if len(res) > 0 && res[0].Equal(net.ParseIP(configs.IP)) {
		return domain
	}

	if domain == "" {
		greenlogger.LogMessagef("Please enter a domain name that redirects to the same IP address you have entered.")
	} else {
		greenlogger.LogMessagef("%v doesn't map to the configured IP address %v , Please enter a valid domain name:", domain, configs.IP)
	}

	var newAddr string
	_, scanErr := fmt.Scanln(&newAddr)
	if scanErr != nil {
		greenlogger.LogError(scanErr, "Problem scanning domain name input")
	}

	return recursivelyEnsureFunctionalDomain(configs, newAddr)
}

// Checks if the entered in IP address is ipv4. If not, it will recurse until it has one.
func recursivelyEnsureIP(addr string) string {
	var ipFromAddr net.IP = net.ParseIP(addr)

	if ipFromAddr.To4() == nil { // If it's nil, convertinig didn't work
		if addr == "" {
			greenlogger.LogMessage("Please enter the outward-facing IP address of this server.")
		} else {
			greenlogger.LogMessage("Error: " + addr + " isn't a valid IPv4 address. Please enter a valid one:")
		}

		var newAddr string
		_, scanErr := fmt.Scanln(&newAddr)

		if scanErr != nil {
			greenlogger.LogError(scanErr, "Problem scanning IP address input")
		}

		return recursivelyEnsureIP(newAddr)
	}

	return ipFromAddr.String()
}

// Waits 10 seconds, then tries to make a connection to its own root. If it cannot, it will fatal.
func EnsureExternalConnectivity() {

	//Waits because sometimes there's a pane in order to give access to wifi on macs especially
	timer := time.NewTimer(10 * time.Second)

	// Wait for the timer channel to trigger
	<-timer.C

	greenlogger.LogMessage("Ensuring remote connectivity to server...")

	// GET the root of the server
	resp, httpErr := http.Get("https://" + constants.CachedConfigs.DomainName)

	if httpErr != nil {
		greenlogger.LogErrorf(httpErr, "Problem sending a GET to %v", "https://"+constants.CachedConfigs.DomainName)
	}

	if resp != nil {
		return
	}

	greenlogger.FatalLogMessage("Unable to externally connect to the server! Make sure all your ports are forwarded right and such things.")
}

// Validates the spreadshet id entered in is valid. If not, recurses until it can return a valid one.
func recursivelyEnsureSpreadsheetID(id string) string {
	if sheet.IsSheetValid(id) {
		sheet.SpreadsheetId = id
		return id
	}

	if id == "" {
		greenlogger.LogMessagef("Please enter a google sheets spreadsheet ID (the part in the url in between d/ and /edit ) that the account your token is associated with can edit.")
	} else {
		greenlogger.LogMessagef("Google Sheets spreadsheet ID %v is invalid, or you don't have permission to access it. Please enter an id of a spreadsheet that will work. It may be that your access token has expired [delete token.json in the runtime directory]", id)
	}
	var newId string
	_, scanErr := fmt.Scanln(&newId)
	if scanErr != nil {
		greenlogger.LogError(scanErr, "Problem scanning spreadsheet ID input")
	}

	return recursivelyEnsureSpreadsheetID(newId)
}

// Ensures that the databases auth.db and users.db exist. If not, it will exit the program.
// Only checks for the files, not their contents. Keeping those in line is up to the maintainer.
func ensureDatabasesExist(configs constants.GeneralConfigs) {
	_, authErr := os.Open(filepath.Join(configs.PathToDatabases, "auth.db"))
	_, usersErr := os.Open(filepath.Join(configs.PathToDatabases, "users.db"))

	if errors.Is(authErr, os.ErrNotExist) || errors.Is(usersErr, os.ErrNotExist) {
		greenlogger.LogMessage("One or both of your essential databases are missing. If you are a member of our organization on github, run")
		greenlogger.LogMessage(`git clone https://github.com/TheGreenMachine/GreenScout-Databases.git in this directory. If not, there are functions to generate your own directories in userDB.go and auth.go !!! THESE DON'T EXIST YET, PLEASE CREATE THEM FUTURE DEVS !!!`)
		os.Exit(1)
	}
}

// Downloads the tba api client for python. If it cannot install with either pip or pip3, it will fatal.
func downloadAPIPackage() {
	c := exec.Command("py", "test.py") //This file legit only has lines of code attempting to import the TBA API, if it fails, it'll throw an exception and the code below will install the API.

	if err := c.Run(); err != nil {
		greenlogger.LogError(err, "woa")
		runnable := exec.Command("pip", "install", "git+https://github.com/TBA-API/tba-api-client-python.git")
		_, execErr := runnable.Output()

		if execErr != nil && !strings.Contains(execErr.Error(), "exit status 1") {
			greenlogger.LogError(execErr, "Problem executing pip install git+https://github.com/TBA-API/tba-api-client-python.git")
			greenlogger.LogMessage("Attempting to run with pip3...")

			runnable = exec.Command("pip3", "install", "git+https://github.com/TBA-API/tba-api-client-python.git")
			_, err := runnable.Output()
			if err != nil && !strings.Contains(err.Error(), "exit status 1") {
				greenlogger.FatalError(err, "Could not install tba-api-client-python with pip or pip3! Please ensure you have pip in your $PATH")
			}
		}
	} else {
		greenlogger.ELogMessage("Gasp")
	}
}

// Runs the configuration routine for custom (non-TBA events)
func configCustomEvent(configs constants.GeneralConfigs) constants.CustomEventConfigs {
	if !configs.CustomEventConfigs.Configured {
		greenlogger.LogMessage("Will your custom event have a schedule? Enter yes if so, anything else if not.")
		var response string
		_, scanErr := fmt.Scanln(&response)

		if scanErr != nil {
			greenlogger.LogError(scanErr, "Problem scanning custom event schedule confirmation")
		}

		configs.CustomEventConfigs.CustomSchedule = response == "yes"
	}

	if configs.CustomEventConfigs.CustomSchedule {
		greenlogger.LogMessagef("Using %s/schedule.json as the match schedule! Please make that it meets your non-TBA event schedule manually.", constants.CachedConfigs.RuntimeDirectory)
	} else {
		schedule.WipeSchedule()
		greenlogger.LogMessage("Not using a schedule.")
	}

	configs.CustomEventConfigs.Configured = true

	return configs.CustomEventConfigs
}
