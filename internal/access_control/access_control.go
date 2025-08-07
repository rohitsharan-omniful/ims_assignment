package access_control

import (
	"context"
	"errors"
	"sync"

	"github.com/gin-gonic/gin"
	constants2 "github.com/omniful/api-gateway/constants"
	"github.com/omniful/ims_rohit/internal/hub"
	"github.com/omniful/ims_rohit/internal/seller"

	"github.com/omniful/go_commons/constants"
	"github.com/omniful/go_commons/jwt/private"
	"github.com/omniful/go_commons/rules/models"
	"github.com/omniful/go_commons/util"
)

type AccessControl struct {
	hubCache    *hub.Cache
	sellerCache *seller.Cache
}

var (
	accessControl *AccessControl
	once          sync.Once
)

func NewAccessControl(hubCache *hub.Cache, sellerCache *seller.Cache) *AccessControl {
	once.Do(func() {
		accessControl = &AccessControl{
			hubCache:    hubCache,
			sellerCache: sellerCache,
		}
	})

	return accessControl
}

func (ac *AccessControl) ValidateAndSetSellerIDs(c *gin.Context, tenantID string, sellerIDsToBeValidated []string) (bool, error) {
	if len(sellerIDsToBeValidated) == 0 {
		return true, setUserSellersInQueryParam(c)
	}

	isSellerValid, err := ac.ValidateSellerIDs(c, tenantID, sellerIDsToBeValidated)
	if err != nil {
		return false, err
	}

	return isSellerValid, setSellerIDsInQueryParam(c, sellerIDsToBeValidated)
}

func (ac *AccessControl) ValidateAndSetHubIDs(c *gin.Context, tenantID string, hubIDsToBeValidated []string) (bool, error) {
	if len(hubIDsToBeValidated) == 0 {
		return true, setUserHubsInQueryParam(c)
	}

	isHubValid, err := ac.ValidateHubIDs(c, tenantID, hubIDsToBeValidated)
	if err != nil {
		return false, err
	}

	return isHubValid, setHubIDsInQueryParam(c, hubIDsToBeValidated)
}

func (ac *AccessControl) ValidateHubIDs(c context.Context, tenantID string, hubIDsToBeValidated []string) (bool, error) {
	if len(hubIDsToBeValidated) == 0 {
		return true, nil
	}

	tenantHubIDs, err := ac.hubCache.GetTenantHubIDs(c, tenantID)
	if err != nil {
		return false, err
	}

	tenantHubs := make(map[string]bool)
	for _, v := range tenantHubIDs {
		tenantHubs[v] = true
	}

	userDetails, ok := c.Value(constants.PrivateUserDetails).(*private.UserDetails)
	if !ok {
		err = errors.New("user details not found in ctx")
		return false, err
	}

	for _, hubIDToBeValidated := range hubIDsToBeValidated {
		if _, present := tenantHubs[hubIDToBeValidated]; !present {
			return false, nil
		}

		if len(hubIDsToBeValidated) > 0 {
			userHasHubAccess, ruleErr := userDetails.RuleGroup.RuleValid(map[string]string{
				"hub_id": hubIDToBeValidated,
			}, []models.Name{models.UserHub})
			if ruleErr != nil {
				return false, ruleErr
			}

			if !userHasHubAccess {
				return false, nil
			}
		}
	}

	return true, nil
}

func (ac *AccessControl) ValidateSellerIDs(c context.Context, tenantID string, sellerIDsToBeValidated []string) (bool, error) {
	if len(sellerIDsToBeValidated) == 0 {
		return true, nil
	}

	tenantSellerIDs, err := ac.sellerCache.GetTenantSellerIDs(c, tenantID)
	if err != nil {
		return false, err
	}

	tenantSellers := make(map[string]bool)
	for _, v := range tenantSellerIDs {
		tenantSellers[v] = true
	}

	userDetails, ok := c.Value(constants.PrivateUserDetails).(*private.UserDetails)
	if !ok {
		err = errors.New("user details not found in ctx")
		return false, err
	}

	for _, sellerIDToBeValidated := range sellerIDsToBeValidated {
		if _, present := tenantSellers[sellerIDToBeValidated]; !present {
			return false, nil
		}

		userHasSellerAccess, ruleErr := userDetails.RuleGroup.RuleValid(map[string]string{
			"seller_id": sellerIDToBeValidated,
		}, []models.Name{models.Seller})
		if ruleErr != nil {
			return false, ruleErr
		}

		if !userHasSellerAccess {
			return false, nil
		}
	}

	return true, nil
}

func setUserSellersInQueryParam(c *gin.Context) error {
	userDetails, ok := c.Value(constants.PrivateUserDetails).(*private.UserDetails)
	if !ok {
		return errors.New("user details not found in ctx")
	}

	userSellers, isAllSellers, err := getUserSellers(userDetails.RuleGroup.Rules)
	if err != nil {
		return err
	}

	if isAllSellers {
		userSellers = make([]string, 0)
	}

	c.Set(constants2.SellerIDs, userSellers)

	return nil
}

func setSellerIDsInQueryParam(c *gin.Context, sellerIDs []string) error {
	userDetails, ok := c.Value(constants.PrivateUserDetails).(*private.UserDetails)
	if !ok {
		return errors.New("user details not found in ctx")
	}

	userSellers, isAllSeller, err := getUserSellers(userDetails.RuleGroup.Rules)
	if err != nil {
		return err
	}

	res := sellerIDs
	if !isAllSeller {
		res = util.Intersection(sellerIDs, userSellers)
	}

	c.Set(constants2.SellerIDs, res)

	return nil
}

func getUserSellers(rules []*models.Rule) (sellerIDs []string, isAllSeller bool, err error) {
	var userSellerRule *models.Rule
	for _, v := range rules {
		if v.Name == models.Seller {
			userSellerRule = v
		}
	}

	if userSellerRule == nil {
		err = errors.New("user hub scope not found")
		return
	}

	for _, v := range userSellerRule.Conditions {
		if v.Operator == models.All {
			isAllSeller = true
			break
		}

		for _, sellerID := range v.Values {
			sellerIDs = append(sellerIDs, sellerID)
		}
	}

	return
}

func setUserHubsInQueryParam(c *gin.Context) error {
	userDetails, ok := c.Value(constants.PrivateUserDetails).(*private.UserDetails)
	if !ok {
		return errors.New("user details not found in ctx")
	}

	userHubs, isAllHubs, err := getUserHubs(userDetails.RuleGroup.Rules)
	if err != nil {
		return err
	}

	if isAllHubs {
		userHubs = make([]string, 0)
	}

	c.Set(constants2.HubIDs, userHubs)

	return nil
}

func setHubIDsInQueryParam(c *gin.Context, hubIDs []string) error {
	userDetails, ok := c.Value(constants.PrivateUserDetails).(*private.UserDetails)
	if !ok {
		return errors.New("user details not found in ctx")
	}

	userHubs, isAllHubs, err := getUserHubs(userDetails.RuleGroup.Rules)
	if err != nil {
		return err
	}

	res := hubIDs
	if !isAllHubs {
		res = util.Intersection(hubIDs, userHubs)
	}

	c.Set(constants2.HubIDs, res)

	return nil
}

func getUserHubs(rules []*models.Rule) (hubIDs []string, isAllHub bool, err error) {
	var userHubRule *models.Rule
	for _, v := range rules {
		if v.Name == models.UserHub {
			userHubRule = v
		}
	}

	if userHubRule == nil {
		err = errors.New("user hub scope not found")
		return
	}

	for _, v := range userHubRule.Conditions {
		if v.Operator == models.All {
			isAllHub = true
			break
		}

		for _, hubID := range v.Values {
			hubIDs = append(hubIDs, hubID)
		}
	}

	return
}
