package api

import (
	"net/http"

	"github.com/hantdev/mitras"
)

var _ mitras.Response = (*publishMessageRes)(nil)

type publishMessageRes struct{}

func (res publishMessageRes) Code() int {
	return http.StatusAccepted
}

func (res publishMessageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res publishMessageRes) Empty() bool {
	return true
}
