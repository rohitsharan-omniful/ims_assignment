package handler

import (
	"context"
	"encoding/json"

	"strconv"

	"github.com/omniful/api-gateway/pkg/redis"
	"github.com/omniful/api-gateway/pkg/serializer"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/pubsub"
	pkgcache "github.com/omniful/go_commons/redis_cache"
	"github.com/omniful/ims_rohit/internal/seller"
)

const (
	SellerCreate   = "seller.create.event"
	SellerUpdate   = "seller.update.event"
	SellerInactive = "seller.inactive.event"
)

type Seller struct {
	ID       uint64 `json:"id"`
	TenantID uint64 `json:"tenant_id"`
}

type InvalidateSellerCache struct {
	sellerCache *seller.Cache
}

func NewInvalidateSellerCacheHandler(ctx context.Context) *InvalidateSellerCache {
	redisClient := pkgcache.NewRedisCacheClient(redis.GetClient().Client, serializer.NewMsgpackSerializer(), config.GetString(ctx, "service.name"))
	sellerCache, cacheErr := seller.NewCache(ctx, redisClient)
	if cacheErr != nil {
		panic(cacheErr)
	}

	return &InvalidateSellerCache{
		sellerCache: sellerCache,
	}
}

func isEventValid(event string) bool {
	if event == SellerCreate {
		return true
	}

	return false
}

func (c *InvalidateSellerCache) Process(ctx context.Context, message *pubsub.Message) error {
	event := message.Headers["event"]
	s := Seller{}

	if !isEventValid(event) {
		return nil
	}

	err := json.Unmarshal(message.Value, &s)
	if err != nil {
		log.Errorf("Error in unmarshalling seller update request: %s", err.Error())
		return nil
	}

	err = c.sellerCache.InvalidateSellerIDs(ctx, strconv.FormatUint(s.TenantID, 10))
	if err != nil {
		return err
	}

	return nil
}
