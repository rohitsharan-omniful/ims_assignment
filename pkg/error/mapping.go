package error

import (
	oerror "github.com/omniful/go_commons/error"
	"github.com/omniful/go_commons/http"
	"github.com/omniful/go_commons/response"
)

var CustomCodeToHttpCodeMapping = map[oerror.Code]http.StatusCode{
	oerror.RateLimitError: http.StatusTooManyRequests,
	oerror.RequestInvalid: http.StatusBadRequest,
}

func Initialize() {
	response.SetCustomErrorMapping(CustomCodeToHttpCodeMapping)
}
