package response

import (
	"github.com/gin-gonic/gin"
	"github.com/omniful/go_commons/http"
	interservice_client "github.com/omniful/go_commons/interservice-client"
	"github.com/omniful/go_commons/response"
)

func (r *Handler) NewErrorResponseByInterServiceError(ctx *gin.Context, error *interservice_client.Error) {
	res := &response.ErrorResponse{
		IsSuccess:  false,
		StatusCode: error.StatusCode.Code(),
		Error: response.Error{
			Message: error.Message,
			Errors:  error.Errors,
		},
	}

	ctx.AbortWithStatusJSON(error.StatusCode.Code(), res)
}

func (r *Handler) NewErrorWithDataResponseByInterServiceError(ctx *gin.Context, error *interservice_client.Error) {
	res := &response.ErrorResponse{
		IsSuccess:  false,
		StatusCode: error.StatusCode.Code(),
		Error: response.Error{
			Message: error.Message,
			Errors:  error.Errors,
			Data:    error.Data,
		},
	}

	ctx.AbortWithStatusJSON(error.StatusCode.Code(), res)
}

func (r *Handler) NewErrorResponseByStatusCode(ctx *gin.Context, statusCode http.StatusCode) {
	res := &response.ErrorResponse{
		IsSuccess:  false,
		StatusCode: statusCode.Code(),
		Error: response.Error{
			Message: statusCode.String(),
		},
	}

	ctx.AbortWithStatusJSON(statusCode.Code(), res)
	ctx.AbortWithStatusJSON(statusCode.Code(), res)
}
