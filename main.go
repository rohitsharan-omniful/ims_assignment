package main

import (
	"context"
	"flag"
	"strings"
	"time"

	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/http"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/shutdown"
	"github.com/omniful/go_commons/worker/configs"
	appinit "github.com/omniful/ims_rohit/init"
	"github.com/omniful/ims_rohit/router"
	"github.com/omniful/ims_rohit/workers"
)

const (
	modeWorker = "worker"
	modeHttp   = "http"
)

func main() {
	// Initialize config
	err := config.Init(time.Second * 10)
	if err != nil {
		log.Panicf("Error while initialising config, err: %v", err)
		panic(err)
	}

	ctx, err := config.TODOContext()
	if err != nil {
		log.Panicf("Error while getting context from config, err: %v", err)
		panic(err)
	}

	appinit.Initialize(ctx)

	var mode, includeGroupArg, excludeGroupArg, ListenerNamesArg string
	flag.StringVar(
		&mode,
		"mode",
		modeHttp,
		"Pass the flag to run in different modes (worker or http)",
	)

	flag.StringVar(
		&includeGroupArg,
		"includeGroups",
		"",
		"Comma-separated list of worker groups to include (default: none)",
	)

	flag.StringVar(
		&excludeGroupArg,
		"excludeGroups",
		"",
		"Comma-separated list of worker groups to exclude (default: none)",
	)

	flag.StringVar(
		&ListenerNamesArg,
		"listenerNames",
		"",
		"Comma-separated list of listener names to start (default: none)",
	)

	flag.Parse()

	server := http.InitializeServer(":8090",
		35*time.Second,
		35*time.Second,
		70*time.Second,
		true,
	)

	switch strings.ToLower(mode) {
	case modeHttp:
		runHttpServer(ctx, server)
	case modeWorker:
		serverConfig := configs.ServerConfig{
			IncludeGroupsArg: includeGroupArg,
			ExcludeGroupsArg: excludeGroupArg,
			ListenerNamesArg: ListenerNamesArg,
		}
		workers.Run(ctx, server, serverConfig)
	default:
		runHttpServer(ctx, server)
	}
}

func runHttpServer(ctx context.Context, server *http.Server) {
	// Initialize middlewares and routes
	engine := router.SetupRouter()
	// Since server embeds *gin.Engine, we can directly use the engine
	*server.Engine = *engine

	log.Infof("Starting server on port" + config.GetString(ctx, "server.port"))

	err := server.StartServer("API Gateway")
	if err != nil {
		log.Errorf(err.Error())
	}

	<-shutdown.GetWaitChannel()
}
