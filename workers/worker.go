package workers

import (
	"context"

	"github.com/omniful/go_commons/http"
	"github.com/omniful/go_commons/worker"
	"github.com/omniful/go_commons/worker/configs"
	"github.com/omniful/go_commons/worker/registry"
)

func Run(
	ctx context.Context,
	httpServer *http.Server,
	serverConfig configs.ServerConfig,
) {
	listenerRegistry := registry.NewRegistry()

	registerDefaultListener(ctx, httpServer, listenerRegistry)
	registerKafkaListeners(ctx, listenerRegistry)

	server := worker.NewServerFromRegistry(listenerRegistry)
	server.RunFromConfig(ctx, serverConfig)
}
