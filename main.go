package main

import (
	_ "embed"
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/httpsrv"
	_ "github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/httpsrv/resource"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/middleware"
	"github.com/mario-imperato/r3ds9-apicommon/linkedservices"
	_ "github.com/mario-imperato/r3ds9-apigtw/rest/api"
	_ "github.com/mario-imperato/r3ds9-apigtw/rest/ui"
	r3ds9MdbApiGtw "github.com/mario-imperato/r3ds9-mongodb/model/r3ds9-apigtw"
	r3ds9MdbVersion "github.com/mario-imperato/r3ds9-mongodb/version"
	"github.com/rs/zerolog/log"

	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

/*
 * tpm-morphia -collection-def-scan-path ./model -out-dir ./model
 */

//go:embed sha.txt
var sha string

//go:embed VERSION
var version string

// appLogo contains the ASCII splash screen
//
//go:embed app-logo.txt
var appLogo []byte

func main() {
	fmt.Println(string(appLogo))
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Sha: %s\n", sha)

	appCfg, err := ReadConfig()
	if nil != err {
		log.Fatal().Err(err).Send()
	}

	log.Info().Interface("config", appCfg).Msg("configuration loaded")
	log.Info().Str("r3ds9_mongodb.ver", r3ds9MdbVersion.VERSION).Msg("initialize stores")
	r3ds9MdbApiGtw.InitStore(0, 0)

	/*
		jc, err := InitGlobalTracer()
		if nil != err {
			log.Fatal().Err(err).Send()
		}
		defer jc.Close()
	*/

	err = linkedservices.InitRegistry(appCfg.App.Services)
	if nil != err {
		log.Fatal().Err(err).Msg("linked services initialization error")
	}

	if appCfg.App.MwRegistry != nil {
		if err := middleware.InitializeHandlerRegistry(appCfg.App.MwRegistry); err != nil {
			log.Fatal().Err(err).Send()
		}
	}

	// shutdownChannel := make(chan os.Signal, 1)
	// signal.Notify(shutdownChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info().Msg("Enabling SIGINT e SIGTERM")
	shutdownChannel := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		shutdownChannel <- fmt.Errorf("signal received: %v", <-c)
	}()

	var wg sync.WaitGroup

	s, err := httpsrv.NewServer(appCfg.App.Http /* , httpsrv.WithListenPort(9090), httpsrv.WithDocumentRoot("/www", "/tmp", false) */)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	if err := s.Start(); err != nil {
		log.Fatal().Err(err).Send()
	}
	defer s.Stop()

	for !s.IsReady() {
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	sig := <-shutdownChannel
	log.Debug().Interface("signal", sig).Msg("got termination signal")

	wg.Wait()
	log.Info().Msg("terminated...")
}

/*
func InitGlobalTracer() (*tracing.Tracer, error) {
	tracer, err := tracing.NewTracer()
	if err != nil {
		return nil, err
	}

	return tracer, err
}
*/
