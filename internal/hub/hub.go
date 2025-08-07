package hub

import (
	"context"
	"errors"
	"github.com/omniful/api-gateway/pkg/redis"
	"github.com/omniful/api-gateway/pkg/serializer"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/constants"
	oerror "github.com/omniful/go_commons/error"
	"github.com/omniful/go_commons/jwt/private"
	"github.com/omniful/go_commons/jwt/public"
	pkgcache "github.com/omniful/go_commons/redis_cache"
	"github.com/omniful/go_commons/rules/models"
)

func ValidateHub(ctx context.Context, hubID string) (isValid bool, err error) {
	tenantID, err := public.GetTenantID(ctx)
	if err != nil {
		return
	}

	redisClient := pkgcache.NewRedisCacheClient(redis.GetClient().Client, serializer.NewMsgpackSerializer(), config.GetString(ctx, "service.name"))
	hubCache, err := NewCache(ctx, redisClient)
	if err != nil {
		return
	}

	tenantHubIDs, err := hubCache.GetTenantHubIDs(ctx, tenantID)
	if err != nil {
		return
	}

	for _, tenantHubID := range tenantHubIDs {
		if tenantHubID == hubID {
			isValid = true
		}
	}

	userDetails, ok := ctx.Value(constants.PrivateUserDetails).(*private.UserDetails)
	if !ok {
		err = errors.New("user details not found in ctx")
		return
	}

	userHasHubAccess, err := userDetails.RuleGroup.RuleValid(map[string]string{
		"hub_id": hubID,
	}, []models.Name{models.UserHub})
	if err != nil {
		return
	}

	isValid = isValid && userHasHubAccess
	return
}

func ValidateAllHubsAccess(ctx context.Context) (hasAllHubsAccess bool, cusErr oerror.CustomError) {
	rules, cusErr := private.GetRuleGroup(ctx)
	if cusErr.Exists() {
		return
	}

	var userHubRule *models.Rule
	for _, v := range rules.Rules {
		if v.Name == models.UserHub {
			userHubRule = v
		}
	}

	if userHubRule == nil {
		cusErr = oerror.NewCustomError(oerror.RulesNotFoundInCtx, "user hub scope not found")
		return
	}

	for _, v := range userHubRule.Conditions {
		if v.Operator == models.All {
			hasAllHubsAccess = true
			return
		}
	}

	return
}
