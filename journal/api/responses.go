package api

import (
	"net/http"

	"github.com/hantdev/mitras"
	"github.com/hantdev/mitras/journal"
)

var _ mitras.Response = (*pageRes)(nil)

type pageRes struct {
	journal.JournalsPage `json:",inline"`
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
