package request

import "time"

type AccessControlRequest struct {
	Url         string
	Body        AccessControlBody   // the request body, should be json serializable
	QueryParams map[string][]string // the request query url params
	Headers     map[string][]string // to set any custom headers, if any
	PathParams  map[string]string
	Timeout     time.Duration
}

type AccessControlBody interface {
	GetHubIDs() []string
	GetSellerIDs() []string
}
