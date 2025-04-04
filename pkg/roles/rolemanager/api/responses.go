package http

import (
	"net/http"

	"github.com/hantdev/mitras"
	"github.com/hantdev/mitras/pkg/roles"
)

var (
	_ mitras.Response = (*createRoleRes)(nil)
	_ mitras.Response = (*listRolesRes)(nil)
	_ mitras.Response = (*viewRoleRes)(nil)
	_ mitras.Response = (*updateRoleRes)(nil)
	_ mitras.Response = (*deleteRoleRes)(nil)
	_ mitras.Response = (*listAvailableActionsRes)(nil)
	_ mitras.Response = (*addRoleActionsRes)(nil)
	_ mitras.Response = (*listRoleActionsRes)(nil)
	_ mitras.Response = (*deleteRoleActionsRes)(nil)
	_ mitras.Response = (*deleteAllRoleActionsRes)(nil)
	_ mitras.Response = (*addRoleMembersRes)(nil)
	_ mitras.Response = (*listRoleMembersRes)(nil)
	_ mitras.Response = (*deleteRoleMembersRes)(nil)
	_ mitras.Response = (*deleteAllRoleMemberRes)(nil)
)

type createRoleRes struct {
	roles.RoleProvision
}

func (res createRoleRes) Code() int {
	return http.StatusCreated
}

func (res createRoleRes) Headers() map[string]string {
	return map[string]string{}
}

func (res createRoleRes) Empty() bool {
	return false
}

type listRolesRes struct {
	roles.RolePage
}

func (res listRolesRes) Code() int {
	return http.StatusOK
}

func (res listRolesRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listRolesRes) Empty() bool {
	return false
}

type listEntityMembersRes struct {
	roles.MembersRolePage
}

func (res listEntityMembersRes) Code() int {
	return http.StatusOK
}

func (res listEntityMembersRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listEntityMembersRes) Empty() bool {
	return false
}

type deleteEntityMembersRes struct{}

func (res deleteEntityMembersRes) Code() int {
	return http.StatusNoContent
}

func (res deleteEntityMembersRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deleteEntityMembersRes) Empty() bool {
	return true
}

type viewRoleRes struct {
	roles.Role
}

func (res viewRoleRes) Code() int {
	return http.StatusOK
}

func (res viewRoleRes) Headers() map[string]string {
	return map[string]string{}
}

func (res viewRoleRes) Empty() bool {
	return false
}

type updateRoleRes struct {
	roles.Role
}

func (res updateRoleRes) Code() int {
	return http.StatusOK
}

func (res updateRoleRes) Headers() map[string]string {
	return map[string]string{}
}

func (res updateRoleRes) Empty() bool {
	return false
}

type deleteRoleRes struct{}

func (res deleteRoleRes) Code() int {
	return http.StatusNoContent
}

func (res deleteRoleRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deleteRoleRes) Empty() bool {
	return true
}

type listAvailableActionsRes struct {
	AvailableActions []string `json:"available_actions"`
}

func (res listAvailableActionsRes) Code() int {
	return http.StatusOK
}

func (res listAvailableActionsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listAvailableActionsRes) Empty() bool {
	return false
}

type addRoleActionsRes struct {
	Actions []string `json:"actions"`
}

func (res addRoleActionsRes) Code() int {
	return http.StatusOK
}

func (res addRoleActionsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res addRoleActionsRes) Empty() bool {
	return false
}

type listRoleActionsRes struct {
	Actions []string `json:"actions"`
}

func (res listRoleActionsRes) Code() int {
	return http.StatusOK
}

func (res listRoleActionsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listRoleActionsRes) Empty() bool {
	return false
}

type deleteRoleActionsRes struct{}

func (res deleteRoleActionsRes) Code() int {
	return http.StatusNoContent
}

func (res deleteRoleActionsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deleteRoleActionsRes) Empty() bool {
	return true
}

type deleteAllRoleActionsRes struct{}

func (res deleteAllRoleActionsRes) Code() int {
	return http.StatusNoContent
}

func (res deleteAllRoleActionsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deleteAllRoleActionsRes) Empty() bool {
	return true
}

type addRoleMembersRes struct {
	Members []string `json:"members"`
}

func (res addRoleMembersRes) Code() int {
	return http.StatusOK
}

func (res addRoleMembersRes) Headers() map[string]string {
	return map[string]string{}
}

func (res addRoleMembersRes) Empty() bool {
	return false
}

type listRoleMembersRes struct {
	roles.MembersPage
}

func (res listRoleMembersRes) Code() int {
	return http.StatusOK
}

func (res listRoleMembersRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listRoleMembersRes) Empty() bool {
	return false
}

type deleteRoleMembersRes struct{}

func (res deleteRoleMembersRes) Code() int {
	return http.StatusNoContent
}

func (res deleteRoleMembersRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deleteRoleMembersRes) Empty() bool {
	return true
}

type deleteAllRoleMemberRes struct{}

func (res deleteAllRoleMemberRes) Code() int {
	return http.StatusNoContent
}

func (res deleteAllRoleMemberRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deleteAllRoleMemberRes) Empty() bool {
	return true
}
