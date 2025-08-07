package hub

import (
	"context"
	"errors"
	"github.com/omniful/api-gateway/constants"
	"github.com/omniful/api-gateway/external_service/wms/private"
	cache2 "github.com/omniful/go_commons/redis_cache"
	"sync"
)

type Cache struct {
	cache     cache2.ICache
	wmsClient private.Client
}

var cache *Cache
var cacheErr error
var cacheOnce sync.Once

func NewCache(ctx context.Context, c cache2.ICache) (*Cache, error) {
	cacheOnce.Do(func() {
		wmsClient, err := private.NewClient(ctx)
		if err != nil {
			cacheErr = err
		}

		cache = &Cache{
			cache:     c,
			wmsClient: wmsClient,
		}
	})

	return cache, cacheErr
}

func (c *Cache) GetTenantHubIDs(ctx context.Context, tenantID string) (hubIDs []string, err error) {
	cacheKey := getTenantHubsCacheKey(ctx, tenantID)
	found, err := c.cache.Get(ctx, cacheKey, &hubIDs)
	if err != nil {
		return
	}

	if found {
		return
	}

	tenantHubs, interSvcErr := c.wmsClient.GetTenantHubs(ctx, tenantID)
	if interSvcErr != nil {
		err = errors.New(interSvcErr.Message)
		return
	}

	for _, hub := range tenantHubs.HubIDs {
		hubIDs = append(hubIDs, hub)
	}

	err = c.setTenantHubIDs(ctx, tenantID, hubIDs)
	if err != nil {
		return
	}
	return
}

func (c *Cache) setTenantHubIDs(ctx context.Context, tenantID string, sellers []string) (err error) {
	cacheKey := getTenantHubsCacheKey(ctx, tenantID)
	_, err = c.cache.Set(ctx, cacheKey, sellers, constants.OneDayExpiration)
	if err != nil {
		return
	}
	return
}

func (c *Cache) InvalidateTenantHubIDs(ctx context.Context, tenantID string) (err error) {
	cacheKey := getTenantHubsCacheKey(ctx, tenantID)
	_, err = c.cache.Unlink(ctx, []string{cacheKey})
	if err != nil {
		return
	}
	return
}

func getTenantHubsCacheKey(ctx context.Context, tenantID string) string {
	return "tenant_hubs_" + tenantID
}
