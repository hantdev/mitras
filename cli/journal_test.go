package cli_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hantdev/mitras/cli"
	"github.com/hantdev/mitras/internal/testsutil"
	"github.com/hantdev/mitras/pkg/errors"
	svcerr "github.com/hantdev/mitras/pkg/errors/service"
	mitrassdk "github.com/hantdev/mitras/pkg/sdk"
	sdkmocks "github.com/hantdev/mitras/pkg/sdk/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var journal = mitrassdk.Journal{
	ID: testsutil.GenerateUUID(&testing.T{}),
}

func TestGetJournalCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	invCmd := cli.NewJournalCmd()
	rootCmd := setFlags(invCmd)

	var page mitrassdk.JournalsPage
	entityType := "group"
	entityId := testsutil.GenerateUUID(t)
	domainId := testsutil.GenerateUUID(t)

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		page          mitrassdk.JournalsPage
		logType       outputLog
		errLogMessage string
	}{
		{
			desc: "get user journal",
			args: []string{
				"user",
				entityId,
				token,
			},
			logType: entityLog,
			page: mitrassdk.JournalsPage{
				Total:    1,
				Offset:   0,
				Limit:    10,
				Journals: []mitrassdk.Journal{journal},
			},
		},
		{
			desc: "get group journal",
			args: []string{
				entityType,
				entityId,
				domainId,
				token,
			},
			logType: entityLog,
			page: mitrassdk.JournalsPage{
				Total:    1,
				Offset:   0,
				Limit:    10,
				Journals: []mitrassdk.Journal{journal},
			},
		},
		{
			desc: "get journal with invalid args",
			args: []string{
				entityType,
				entityId,
				token,
				domainId,
				extraArg,
			},
			logType: usageLog,
		},
		{
			desc: "get journal with invalid token",
			args: []string{
				entityType,
				entityId,
				domainId,
				invalidToken,
			},
			logType:       errLog,
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("Journal", tc.args[0], tc.args[1], "", mock.Anything, tc.args[2]).Return(tc.page, tc.sdkErr)
			if tc.args[0] != "user" {
				sdkCall = sdkMock.On("Journal", tc.args[0], tc.args[1], tc.args[2], mock.Anything, tc.args[3]).Return(tc.page, tc.sdkErr)
			}
			out := executeCommand(t, rootCmd, append([]string{getCmd}, tc.args...)...)

			switch tc.logType {
			case entityLog:
				err := json.Unmarshal([]byte(out), &page)
				assert.Nil(t, err)
				assert.Equal(t, tc.page, page, fmt.Sprintf("%v unexpected response, expected: %v, got: %v", tc.desc, tc.page, page))
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			}
			sdkCall.Unset()
		})
	}
}
