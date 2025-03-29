package pats

import (
	"net/http"

	"github.com/hantdev/mitras"
	"github.com/hantdev/mitras/auth"
)

var (
	_ mitras.Response = (*createPatRes)(nil)
	_ mitras.Response = (*retrievePatRes)(nil)
	_ mitras.Response = (*updatePatNameRes)(nil)
	_ mitras.Response = (*updatePatDescriptionRes)(nil)
	_ mitras.Response = (*deletePatRes)(nil)
	_ mitras.Response = (*resetPatSecretRes)(nil)
	_ mitras.Response = (*revokePatSecretRes)(nil)
	_ mitras.Response = (*scopeRes)(nil)
	_ mitras.Response = (*clearAllRes)(nil)
)

type createPatRes struct {
	auth.PAT `json:",inline"`
}

func (res createPatRes) Code() int {
	return http.StatusCreated
}

func (res createPatRes) Headers() map[string]string {
	return map[string]string{}
}

func (res createPatRes) Empty() bool {
	return false
}

type retrievePatRes struct {
	auth.PAT `json:",inline"`
}

func (res retrievePatRes) Code() int {
	return http.StatusOK
}

func (res retrievePatRes) Headers() map[string]string {
	return map[string]string{}
}

func (res retrievePatRes) Empty() bool {
	return false
}

type updatePatNameRes struct {
	auth.PAT `json:",inline"`
}

func (res updatePatNameRes) Code() int {
	return http.StatusAccepted
}

func (res updatePatNameRes) Headers() map[string]string {
	return map[string]string{}
}

func (res updatePatNameRes) Empty() bool {
	return false
}

type updatePatDescriptionRes struct {
	auth.PAT `json:",inline"`
}

func (res updatePatDescriptionRes) Code() int {
	return http.StatusAccepted
}

func (res updatePatDescriptionRes) Headers() map[string]string {
	return map[string]string{}
}

func (res updatePatDescriptionRes) Empty() bool {
	return false
}

type listPatsRes struct {
	auth.PATSPage `json:",inline"`
}

func (res listPatsRes) Code() int {
	return http.StatusOK
}

func (res listPatsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listPatsRes) Empty() bool {
	return false
}

type deletePatRes struct{}

func (res deletePatRes) Code() int {
	return http.StatusNoContent
}

func (res deletePatRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deletePatRes) Empty() bool {
	return true
}

type resetPatSecretRes struct {
	auth.PAT `json:",inline"`
}

func (res resetPatSecretRes) Code() int {
	return http.StatusOK
}

func (res resetPatSecretRes) Headers() map[string]string {
	return map[string]string{}
}

func (res resetPatSecretRes) Empty() bool {
	return false
}

type revokePatSecretRes struct{}

func (res revokePatSecretRes) Code() int {
	return http.StatusNoContent
}

func (res revokePatSecretRes) Headers() map[string]string {
	return map[string]string{}
}

func (res revokePatSecretRes) Empty() bool {
	return true
}

type scopeRes struct{}

func (res scopeRes) Code() int {
	return http.StatusOK
}

func (res scopeRes) Headers() map[string]string {
	return map[string]string{}
}

func (res scopeRes) Empty() bool {
	return true
}

type clearAllRes struct{}

func (res clearAllRes) Code() int {
	return http.StatusOK
}

func (res clearAllRes) Headers() map[string]string {
	return map[string]string{}
}

func (res clearAllRes) Empty() bool {
	return true
}

type listScopeRes struct {
	auth.ScopesPage `json:",inline"`
}

func (res listScopeRes) Code() int {
	return http.StatusOK
}

func (res listScopeRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listScopeRes) Empty() bool {
	return false
}
