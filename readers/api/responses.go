package api

import (
	"net/http"

	"github.com/hantdev/mitras"
	"github.com/hantdev/mitras/readers"
)

var _ mitras.Response = (*pageRes)(nil)

type pageRes struct {
	readers.PageMetadata
	Total    uint64            `json:"total"`
	Messages []readers.Message `json:"messages,omitempty"`
}

func (res pageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res pageRes) Code() int {
	return http.StatusOK
}

func (res pageRes) Empty() bool {
	return false
}
