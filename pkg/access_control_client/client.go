package access_control_client

import (
	"context"

	constants2 "github.com/omniful/api-gateway/constants"
	"github.com/omniful/ims_rohit/internal/access_control"

	"github.com/omniful/api-gateway/pkg/request"
	"github.com/omniful/go_commons/http"
	interserviceclient "github.com/omniful/go_commons/interservice-client"
	"github.com/omniful/go_commons/jwt/public"
	"github.com/omniful/go_commons/util"
)

type Client struct {
	*interserviceclient.Client
	*access_control.AccessControl
	enforceControl bool
}

func NewClient(client *interserviceclient.Client, accessControl *access_control.AccessControl, enforceControl bool) *Client {
	return &Client{
		Client:         client,
		AccessControl:  accessControl,
		enforceControl: enforceControl,
	}
}

// ExecuteWithAccessControl - To be used only in API Gateway
func (c *Client) ExecuteWithAccessControl(
	ctx context.Context,
	method http.APIMethod,
	request *http.Request,
	data interface{},
) (*interserviceclient.InterSvcResponse, *interserviceclient.Error) {
	err := c.validateRequest(ctx, request, c.enforceControl)
	if err != nil {
		return nil, err
	}

	res, err := c.Execute(ctx, method, &http.Request{
		Url:         request.Url,
		Timeout:     request.Timeout,
		QueryParams: request.QueryParams,
		PathParams:  request.PathParams,
		Headers:     request.Headers,
		Body:        request.Body,
	}, data)
	return res, err
}

func (c *Client) validateRequest(ctx context.Context, req *http.Request, enforceControl bool) (interSvcErr *interserviceclient.Error) {
	if req.Body == nil {
		return
	}

	tenantID, err := public.GetTenantID(ctx)
	if err != nil {
		interSvcErr = &interserviceclient.Error{
			Message:    err.Error(),
			Errors:     nil,
			StatusCode: 500,
		}
	}

	val, isValid := req.Body.(request.AccessControlBody)
	if !isValid && enforceControl {
		interSvcErr = &interserviceclient.Error{
			Message:    constants2.InvalidRequest,
			Errors:     nil,
			StatusCode: 500,
		}
		return
	}

	if !isValid {
		return
	}

	isHubValid, err := c.ValidateHubIDs(ctx, tenantID, util.DeduplicateSlice(val.GetHubIDs()))
	if err != nil || !isHubValid {
		interSvcErr = &interserviceclient.Error{
			Message:    constants2.HUbNotValid,
			Errors:     nil,
			StatusCode: 500,
		}
		return
	}

	isSellerValid, err := c.ValidateSellerIDs(ctx, tenantID, util.DeduplicateSlice(val.GetSellerIDs()))
	if err != nil || !isSellerValid {
		interSvcErr = &interserviceclient.Error{
			Message:    constants2.SellerNotValid,
			Errors:     nil,
			StatusCode: 500,
		}
		return
	}

	return
}
