package utils

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	constants2 "github.com/omniful/api-gateway/constants"
	"github.com/omniful/go_commons/constants"
	error2 "github.com/omniful/go_commons/error"
	"github.com/omniful/go_commons/jwt/private"
)

type DateType string

const (
	FromDate DateType = "from"
	ToDate   DateType = "to"
)

var TMSPublicHeaderKeys = map[string]string{
	constants2.Authorization:        constants2.Authorization,
	constants2.XOmnifulPlatform:     constants2.XOmnifulPlatform,
	constants2.XOmnifulVersion:      constants2.XOmnifulVersion,
	constants2.XOmnifulPlatformName: constants2.XOmnifulPlatformName,
}

func GetDefaultHeaders(ctx context.Context) (headers map[string][]string) {
	headers = make(map[string][]string, 0)
	val := ctx.Value(constants.JWTUserDetails).(string)
	headers[constants.JWTHeader] = []string{val}
	requestID := ctx.Value(constants.HeaderXOmnifulRequestID).(string)
	headers[constants.HeaderXOmnifulRequestID] = []string{requestID}
	return
}

func GetDefaultQueryParams(ctx context.Context) map[string][]string {
	queryParams := make(map[string][]string)

	if hubIDs, ok := ctx.Value(constants2.HubIDs).([]string); ok {
		queryParams[constants2.HubIDs] = hubIDs
	}

	if sellerIDs, ok := ctx.Value(constants2.SellerIDs).([]string); ok {
		queryParams[constants2.SellerIDs] = sellerIDs
	}

	return queryParams
}

func GetDefaultFleetHeaders(tokenString string) (headers map[string][]string, err error) {
	headers = make(map[string][]string)
	headers[constants2.Authorization] = []string{constants2.Bearer + " " + tokenString}
	return
}

func GetDefaultTmsPublicHeaders(ctx context.Context) (headers map[string][]string) {
	headers = make(map[string][]string)

	if c, ok := ctx.(*gin.Context); ok {
		for key, headerKey := range TMSPublicHeaderKeys {
			if value := c.Request.Header.Get(headerKey); len(value) > 0 {
				headers[key] = []string{value}
			}
		}
	}
	return
}

func GetRequestIDHeader(ctx context.Context) (headers map[string][]string) {
	headers = make(map[string][]string, 0)

	if requestID, ok := ctx.Value(constants.HeaderXOmnifulRequestID).(string); ok {
		headers[constants.HeaderXOmnifulRequestID] = []string{requestID}
	}

	return
}

func FormatDateWithLayout(ctx context.Context, date string, layout string) string {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return date
	}
	return t.Format(layout)
}

type UserDetails struct {
	UserName   string
	UserID     string
	UserEmail  string
	TenantName string
	TenantID   string
}

func CtxInfo(ctx context.Context, timezone ...string) (userDetails UserDetails, cusErr error2.CustomError) {
	userID, cusErr := private.GetUserID(ctx)
	if cusErr.Exists() {
		return
	}
	userDetails.UserID = userID

	tenantID, cusErr := private.GetTenantID(ctx)
	if cusErr.Exists() {
		return
	}
	userDetails.TenantID = tenantID

	userName, cusErr := private.GetUserName(ctx)
	if cusErr.Exists() {
		return
	}
	userDetails.UserName = userName

	tenantName, cusErr := private.GetTenantName(ctx)
	if cusErr.Exists() {
		return
	}
	userDetails.TenantName = tenantName

	userEmail, cusErr := private.GetUserEmail(ctx)
	if cusErr.Exists() {
		return
	}
	userDetails.UserEmail = userEmail

	if len(timezone) > 0 {
		_, err := time.LoadLocation(timezone[0])
		if err != nil {
			cusErr = error2.RequestInvalidError("invalid timezone type")
			return
		}
	}

	return
}

type Meta struct {
	CurrentPage int64  `json:"current_page"`
	PerPage     int64  `json:"per_page"`
	LastPage    int64  `json:"last_page"`
	Total       uint64 `json:"total"`
}

func CreateBatches(array []string, batchSize int) [][]string {
	batches := make([][]string, 0, len(array)/batchSize+1)
	for i := 0; i < len(array); i += batchSize {
		end := i + batchSize
		if end > len(array) {
			end = len(array)
		}
		batches = append(batches, array[i:end])
	}

	return batches
}

func GetRequestIDHeaders(ctx context.Context) (headers map[string][]string) {
	headers = make(map[string][]string, 0)

	requestID := ctx.Value(constants.HeaderXOmnifulRequestID).(string)
	headers[constants.HeaderXOmnifulRequestID] = []string{requestID}

	if correlationID, ok := ctx.Value(constants.HeaderXOmnifulCorrelationID).(string); ok && len(correlationID) > 0 {
		headers[constants.HeaderXOmnifulCorrelationID] = []string{correlationID}
	}

	return
}
