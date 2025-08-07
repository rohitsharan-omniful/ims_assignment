package seller

import (
	"context"
	"errors"
	"github.com/omniful/api-gateway/pkg/redis"
	"github.com/omniful/api-gateway/pkg/serializer"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/constants"
	"github.com/omniful/go_commons/jwt/private"
	"github.com/omniful/go_commons/jwt/public"
	pkgcache "github.com/omniful/go_commons/redis_cache"
	"github.com/omniful/go_commons/rules/models"
)

func ValidateSeller(ctx context.Context, sellerID string) (isValid bool, err error) {
	tenantID, err := public.GetTenantID(ctx)
	if err != nil {
		return
	}

	redisClient := pkgcache.NewRedisCacheClient(redis.GetClient().Client, serializer.NewMsgpackSerializer(), config.GetString(ctx, "service.name"))
	sellerCache, err := NewCache(ctx, redisClient)
	if err != nil {
		return
	}

	sellers, err := sellerCache.GetTenantSellerIDs(ctx, tenantID)
	if err != nil {
		return
	}

	for _, seller := range sellers {
		if seller == sellerID {
			isValid = true
		}
	}

	userDetails, ok := ctx.Value(constants.PrivateUserDetails).(*private.UserDetails)
	if !ok {
		err = errors.New("user details not found in ctx")
		return
	}

	userHasSellerAccess, err := userDetails.RuleGroup.RuleValid(map[string]string{
		"seller_id": sellerID,
	}, []models.Name{models.Seller})
	if err != nil {
		return
	}

	isValid = isValid && userHasSellerAccess
	return
}
