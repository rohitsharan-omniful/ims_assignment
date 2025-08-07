package handler

import (
	"context"
	"encoding/json"

	"github.com/omniful/api-gateway/pkg/redis"
	"github.com/omniful/api-gateway/pkg/serializer"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/pubsub"
	pkgcache "github.com/omniful/go_commons/redis_cache"
	"github.com/omniful/ims_rohit/internal/hub"
)

const (
	HubUpdate = "hubs.update.event"
)

type Hub struct {
	ID       uint64 `json:"id"`
	TenantID string `json:"tenant_id"`
}

type InvalidateHubCache struct {
	hubCache *hub.Cache
}

func NewInvalidateHubCacheHandler(ctx context.Context) *InvalidateHubCache {
	redisClient := pkgcache.NewRedisCacheClient(redis.GetClient().Client, serializer.NewMsgpackSerializer(), config.GetString(ctx, "service.name"))
	hubCache, cacheErr := hub.NewCache(ctx, redisClient)
	if cacheErr != nil {
		panic(cacheErr)
	}

	return &InvalidateHubCache{
		hubCache: hubCache,
	}
}

func isHubEventValid(event string) bool {
	return event == HubUpdate
}

func (c *InvalidateHubCache) Process(ctx context.Context, message *pubsub.Message) error {
	event := message.Headers["event"]
	s := Hub{}

	if !isHubEventValid(event) {
		return nil
	}

	err := json.Unmarshal(message.Value, &s)
	if err != nil {
		log.Errorf("Error in unmarshalling hub update request: %s", err.Error())
		return nil
	}

	err = c.hubCache.InvalidateTenantHubIDs(ctx, s.TenantID)
	if err != nil {
		return err
	}

	return nil
}
