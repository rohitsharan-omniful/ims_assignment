package workers

import (
	"context"

	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/http"
	"github.com/omniful/go_commons/pubsub"
	"github.com/omniful/go_commons/worker/configs"
	"github.com/omniful/go_commons/worker/registry"
	"github.com/omniful/ims_rohit/workers/handler"
)

func registerDefaultListener(ctx context.Context, httpServer *http.Server, registry *registry.Registry) {
	registry.RegisterHTTPListenerConfig(httpServer, config.GetString(ctx, "service.name"))
}

func registerKafkaListeners(ctx context.Context, registry *registry.Registry) {
	registry.RegisterKafkaListenerConfig(
		ctx,
		"sellerUpdate",
		func(handlerCtx context.Context, config configs.KafkaConsumerConfig) pubsub.IPubSubMessageHandler {
			return handler.NewInvalidateSellerCacheHandler(ctx)
		},
	)

	registry.RegisterKafkaListenerConfig(
		ctx,
		"hubUpdate",
		func(handlerCtx context.Context, config configs.KafkaConsumerConfig) pubsub.IPubSubMessageHandler {
			return handler.NewInvalidateHubCacheHandler(ctx)
		},
	)
}
