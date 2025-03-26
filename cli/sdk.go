package cli

import mitrassdk "github.com/hantdev/mitras/pkg/sdk"

// Keep SDK handle in global var.
var sdk mitrassdk.SDK

// SetSDK sets supermq SDK instance.
func SetSDK(s mitrassdk.SDK) {
	sdk = s
}