package cli

import smqsdk "github.com/hantdev/mitras/pkg/sdk"

// Keep SDK handle in global var.
var sdk smqsdk.SDK

// SetSDK sets mitras SDK instance.
func SetSDK(s smqsdk.SDK) {
	sdk = s
}
