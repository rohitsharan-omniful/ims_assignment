package seller

import (
	"context"
	"errors"
	"github.com/omniful/api-gateway/constants"
	"github.com/omniful/api-gateway/external_service/tenant/private"
	cache2 "github.com/omniful/go_commons/redis_cache"
	"sync"
)

type Cache struct {
	cache        cache2.ICache
	tenantClient private.Client
}

var cache *Cache
var cacheErr error
var cacheOnce sync.Once

func NewCache(ctx context.Context, c cache2.ICache) (*Cache, error) {
	cacheOnce.Do(func() {
		tenantSvcClient, err := private.NewClient(ctx)
		if err != nil {
			cacheErr = err
		}

		cache = &Cache{
			cache:        c,
			tenantClient: tenantSvcClient,
		}
	})

	return cache, cacheErr
}

func (c *Cache) GetTenantSellerIDs(ctx context.Context, tenantID string) (sellers []string, err error) {
	cacheKey := getTenantSellersCacheKey(ctx, tenantID)
	found, err := c.cache.Get(ctx, cacheKey, &sellers)
	if err != nil {
		return
	}

	if found {
		return
	}

	tenantSellers, interSvcErr := c.tenantClient.GetTenantSellers(ctx, tenantID, map[string][]string{})
	if interSvcErr != nil {
		err = errors.New(interSvcErr.Message)
		return
	}

	for _, v := range *tenantSellers {
		sellers = append(sellers, v.ID)
	}

	err = c.setTenantSellers(ctx, tenantID, sellers)
	if err != nil {
		return
	}
	return
}

func (c *Cache) setTenantSellers(ctx context.Context, tenantID string, sellers []string) (err error) {
	cacheKey := getTenantSellersCacheKey(ctx, tenantID)
	_, err = c.cache.Set(ctx, cacheKey, sellers, constants.OneDayExpiration)
	if err != nil {
		return
	}
	return
}

func (c *Cache) InvalidateSellerIDs(ctx context.Context, sellerID string) (err error) {
	cacheKey := getTenantSellersCacheKey(ctx, sellerID)
	_, err = c.cache.Unlink(ctx, []string{cacheKey})
	if err != nil {
		return
	}
	return
}

func getTenantSellersCacheKey(ctx context.Context, tenantID string) string {
	return "tenant_sellers_" + tenantID
}
