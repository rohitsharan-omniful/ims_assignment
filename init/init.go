package appinit

import (
	"context"
	"time"

	"github.com/omniful/api-gateway/pkg/redis"
	validator "github.com/omniful/api-gateway/pkg/validate"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/newrelic"
	oredis "github.com/omniful/go_commons/redis"
	"github.com/omniful/ims_rohit/pkg/error"
)

func Initialize(ctx context.Context) {
	initializeLog(ctx)
	initializeNewrelic(ctx)
	initializeRedis(ctx)
	validator.Set()
	error.Initialize()
}

// Initialize logging
func initializeLog(ctx context.Context) {
	err := log.InitializeLogger(
		log.Formatter(config.GetString(ctx, "log.format")),
		log.Level(config.GetString(ctx, "log.level")),
	)
	if err != nil {
		log.WithError(err).Panic("unable to initialise log")
	}

}

// Initialize Newrelic
func initializeNewrelic(ctx context.Context) {
	newrelic.Initialize(&newrelic.Options{
		Name:              config.GetString(ctx, "newrelic.appName"),
		License:           config.GetString(ctx, "newrelic.licence"),
		Enabled:           config.GetBool(ctx, "newrelic.enabled"),
		DistributedTracer: config.GetBool(ctx, "newrelic.distributedTracer"),
	})
	log.InfofWithContext(ctx, "Initialized New Relic")
}

// Initialize Redis
func initializeRedis(ctx context.Context) {
	r := oredis.NewClient(&oredis.Config{
		ClusterMode:  config.GetBool(ctx, "redis.clusterMode"),
		Hosts:        config.GetStringSlice(ctx, "redis.hosts"),
		DB:           config.GetUint(ctx, "redis.db"),
		DialTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	})
	log.InfofWithContext(ctx, "Initialized Redis Client")
	redis.SetClient(r)
}
