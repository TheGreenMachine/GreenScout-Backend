package main

import (
	"GreenScoutBackend/internal"
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
	internal.InitLogFile()

	/// Setup
	isSetup := slices.Contains(os.Args, "setup")
	publicHosting := false //Allows setup to bypass ip and domain validation to run localhost
	serveTLS := false      //switches it between https and http [true for https]
	updateDB := false
	httpPort := ":8080"
	httpsPort := ":8443"

	if internal.IsSudo() {
		if isSetup {
			internal.FatalLogMessage("If you are running in setup mode, please run without sudo!")
		}
		httpPort = ":80"
		httpsPort = ":443"
	}

	/// Running mode
	if slices.Contains(os.Args, "prod") {
		if slices.Contains(os.Args, "test") {
			internal.FatalLogMessage("Use only one of 'prod' or 'test'!!")
		}

		publicHosting = true
		serveTLS = internal.EnableHttps
		updateDB = false
	}

	internal.TotalSetup(publicHosting)

	internal.WriteConditionalFormatting()
	if isSetup { // Exit if only in setup mode
		os.Exit(1)
	}

	// Init DBs
	internal.InitScoutDB()
	internal.InitAuthDB()
	internal.InitUserDB()

	internal.StoreTeams()

	// Write all match numbers to the sheet with a 1 minute cooldown to avoid rate limiting
	if slices.Contains(os.Args, "matches") {
		var usingRemainder bool = false

		matches := internal.GetNumMatches()

		blocks := matches / 50

		var remainder int
		if remainder = matches % 50; remainder > 0 { // Remainder
			usingRemainder = true
		}

		for i := 1; i <= blocks*50; i += 50 {
			internal.FillMatches(i, i+49)
			time.Sleep(1 * time.Minute)
		}

		if usingRemainder {
			initial := blocks * 50
			if initial == 0 {
				initial++
			}
			internal.FillMatches(initial, initial+remainder)
		}
	}

	// get server
	jSrv := internal.SetupServer()

	// ACME autocert with letsEncrypt
	var serverManager *autocert.Manager
	if publicHosting {
		serverManager = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(internal.CachedConfigs.DomainName),
			Cache:      autocert.DirCache(internal.CachedConfigs.CertsDirectory), // This may not be the... wisest choice. Anyone in the future, feel free to fix.
		}
		jSrv.TLSConfig = &tls.Config{GetCertificate: serverManager.GetCertificate}

		go func() {
			// HTTP redirect to HTTPS server
			h := serverManager.HTTPHandler(nil)
			internal.FatalError(http.ListenAndServe(httpPort, h), "http.ListenAndServe() failed")
		}()

	}

	if updateDB {
		// Daily commit + push
		cronManager := cron.New()
		_, cronErr := cronManager.AddFunc("@midnight", internal.CommitAndPushDBs)
		if cronErr != nil {
			internal.FatalError(cronErr, "Problem assigning commit and push task to cron")
		}
		cronManager.Start()
	}

	go func() {
		if serveTLS {
			crtPath := ""
			keyPath := ""
			if !publicHosting {
				// Local keys
				crtPath = filepath.Join(internal.CachedConfigs.RuntimeDirectory, "localhost.crt")
				keyPath = filepath.Join(internal.CachedConfigs.RuntimeDirectory, "localhost.key")
			}

			jSrv.Addr = httpsPort
			err := jSrv.ListenAndServeTLS(crtPath, keyPath)
			if err != nil {
				internal.FatalError(err, "jSrv.ListendAndServeTLS() failed")
			}

		} else {
			jSrv.Addr = httpPort
			err := jSrv.ListenAndServe()
			if err != nil {
				internal.FatalError(err, "jSrv.ListendAndServe() failed")
			}
		}
	}()

	if publicHosting {
		internal.EnsureExternalConnectivity()
	}

	internal.LogMessage("Server Successfully Set Up! [ctrl+c to cancel]")

	go internal.RunServerLoop()

	/// Graceful shutdown

	// Listen for termination signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-signalCh

	// no need to os.exit, since the main thread exits here all the goroutines will shut down
}
