package api

import (
	"net/http"

	"github.com/hantdev/mitras"
	"github.com/hantdev/mitras/journal"
)

var (
	_ mitras.Response = (*pageRes)(nil)
	_ mitras.Response = (*clientTelemetryRes)(nil)
)

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

type clientTelemetryRes struct {
	journal.ClientTelemetry `json:",inline"`
}

func (res clientTelemetryRes) Headers() map[string]string {
	return map[string]string{}
}

func (res clientTelemetryRes) Code() int {
	return http.StatusOK
}

func (res clientTelemetryRes) Empty() bool {
	return false
}
