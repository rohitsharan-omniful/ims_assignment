package response

import (
	"github.com/gin-gonic/gin"

	"github.com/omniful/go_commons/http"
	"github.com/omniful/go_commons/jwt/public"
	"github.com/omniful/go_commons/response"
	"github.com/omniful/go_commons/util"
	"github.com/omniful/ims_rohit/internal/access_control"
	"github.com/omniful/ims_rohit/internal/hub"
	"github.com/omniful/ims_rohit/internal/seller"
)

type AccessControlData interface {
	GetHubIDs() []string
	GetSellerIDs() []string
}

type AccessControlSuccessResponse struct {
	IsSuccess  bool        `json:"is_success"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data"`
	Meta       interface{} `json:"meta"`
}

// Handler holds the caches needed for access control validation
type Handler struct {
	accessControl *access_control.AccessControl
}

// NewResponseHandler creates a new ResponseHandler with initialized caches
func NewResponseHandler(hubCache *hub.Cache, sellerCache *seller.Cache) *Handler {
	return &Handler{
		accessControl: access_control.NewAccessControl(hubCache, sellerCache),
	}
}

func (r *Handler) NewAccessControlSuccessResponse(ctx *gin.Context, data AccessControlData) {
	isValid, err := r.validateResponse(ctx, data)
	if err != nil || !isValid {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError.Code(), err)
		return
	}

	res := &response.SuccessResponse{
		IsSuccess:  true,
		StatusCode: http.StatusOK.Code(),
		Data:       data,
	}

	ctx.AbortWithStatusJSON(http.StatusOK.Code(), res)
	return
}

func (r *Handler) NewAccessControlSuccessResponseWithMeta(ctx *gin.Context, data AccessControlData, meta interface{}) {
	isValid, err := r.validateResponse(ctx, data)
	if err != nil || !isValid {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError.Code(), err)
		return
	}

	res := &response.SuccessResponse{
		IsSuccess:  true,
		StatusCode: http.StatusOK.Code(),
		Data:       data,
		Meta:       meta,
	}

	ctx.AbortWithStatusJSON(http.StatusOK.Code(), res)
	return
}

func (r *Handler) validateResponse(ctx *gin.Context, data AccessControlData) (bool, error) {
	tenantID, err := public.GetTenantID(ctx)
	if err != nil {
		return false, err
	}

	isValidHubIDs, err := r.accessControl.ValidateHubIDs(ctx, tenantID, util.DeduplicateSlice(data.GetHubIDs()))
	if err != nil {
		return false, err
	}

	isValidSellerIDs, err := r.accessControl.ValidateSellerIDs(ctx, tenantID, util.DeduplicateSlice(data.GetSellerIDs()))
	if err != nil {
		return false, err
	}

	return isValidHubIDs && isValidSellerIDs, nil
}
