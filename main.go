package main

import (
	"GreenScoutBackend/constants"
	filemanager "GreenScoutBackend/fileManager"
	greenlogger "GreenScoutBackend/greenLogger"
	"GreenScoutBackend/lib"
	"GreenScoutBackend/schedule"
	"GreenScoutBackend/server"
	"GreenScoutBackend/setup"
	"GreenScoutBackend/sheet"
	"GreenScoutBackend/userDB"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	// Initialize log file
	greenlogger.InitLogFile()

	/// Setup
	isSetup := slices.Contains(os.Args, "setup")
	publicHosting := true //Allows setup to bypass ip and domain validation to run localhost
	serveTLS := false     //switches it between https and http [true for https]
	updateDB := false

	if isSetup && filemanager.IsSudo() {
		greenlogger.FatalLogMessage("If you are running in setup mode, please run without sudo!")
	}

	/// Running mode
	if slices.Contains(os.Args, "prod") {
		if slices.Contains(os.Args, "test") {
			greenlogger.FatalLogMessage("Use only one of 'prod' or 'test'!!")
		}

		publicHosting = true
		serveTLS = constants.EnableHttps
		updateDB = false
	}

	setup.TotalSetup(publicHosting)

	sheet.WriteConditionalFormatting()
	if isSetup { // Exit if only in setup mode
		os.Exit(1)
	}

	// Init DBs
	schedule.InitScoutDB()
	userDB.InitAuthDB()
	userDB.InitUserDB()

	lib.StoreTeams()

	// Write all match numbers to the sheet with a 1 minute cooldown to avoid rate limiting
	if slices.Contains(os.Args, "matches") {
		var usingRemainder bool = false

		matches := lib.GetNumMatches()

		blocks := matches / 50

		var remainder int
		if remainder = matches % 50; remainder > 0 { // Remainder
			usingRemainder = true
		}

		for i := 1; i <= blocks*50; i += 50 {
			sheet.FillMatches(i, i+49)
			time.Sleep(1 * time.Minute)
		}

		if usingRemainder {
			initial := blocks * 50
			if initial == 0 {
				initial++
			}
			sheet.FillMatches(initial, initial+remainder)
		}
	}

	// get server
	jSrv := server.SetupServer()

	// ACME autocert with letsEncrypt
	var serverManager *autocert.Manager
	if publicHosting {
		serverManager = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(constants.CachedConfigs.DomainName),
			Cache:      autocert.DirCache(constants.CachedConfigs.CertsDirectory), // This may not be the... wisest choice. Anyone in the future, feel free to fix.
		}
		jSrv.TLSConfig = &tls.Config{GetCertificate: serverManager.GetCertificate}

		go func() {
			// HTTP redirect to HTTPS server
			h := serverManager.HTTPHandler(nil)
			greenlogger.FatalError(http.ListenAndServe(":http", h), "http.ListenAndServe() failed")
		}()

	}

	if updateDB {
		// Daily commit + push
		cronManager := cron.New()
		_, cronErr := cronManager.AddFunc("@midnight", userDB.CommitAndPushDBs)
		if cronErr != nil {
			greenlogger.FatalError(cronErr, "Problem assigning commit and push task to cron")
		}
		cronManager.Start()
	}

	go func() {
		if serveTLS {
			crtPath := ""
			keyPath := ""
			if !publicHosting {
				// Local keys
				crtPath = filepath.Join(constants.CachedConfigs.RuntimeDirectory, "server.crt")
				keyPath = filepath.Join(constants.CachedConfigs.RuntimeDirectory, "server.key")
			}

			jSrv.Addr = ":8443" //https
			err := jSrv.ListenAndServeTLS(crtPath, keyPath)
			if err != nil {
				greenlogger.FatalError(err, "jSrv.ListendAndServeTLS() failed")
			}

		} else {
			jSrv.Addr = ":8080" //http
			err := jSrv.ListenAndServe()
			if err != nil {
				greenlogger.FatalError(err, "jSrv.ListendAndServe() failed")
			}
		}
	}()

	if publicHosting {
		setup.EnsureExternalConnectivity()
	}

	greenlogger.LogMessage("Server Successfully Set Up! [ctrl+c to cancel]")
	if constants.CachedConfigs.SlackConfigs.UsingSlack {
		greenlogger.NotifyOnline(true)
	}

	go server.RunServerLoop()

	/// Graceful shutdown

	// Listen for termination signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-signalCh
	if constants.CachedConfigs.SlackConfigs.UsingSlack {
		greenlogger.NotifyOnline(false)
	}

	// no need to os.exit, since the main thread exits here all the goroutines will shut down
}
