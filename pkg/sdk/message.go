package sdk

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	apiutil "github.com/hantdev/mitras/api/http/util"
	"github.com/hantdev/mitras/pkg/errors"
)

const channelParts = 2

func (sdk mgSDK) SendMessage(ctx context.Context, chanName, msg, key string) errors.SDKError {
	chanNameParts := strings.SplitN(chanName, ".", channelParts)
	chanID := chanNameParts[0]
	subtopicPart := ""
	if len(chanNameParts) == channelParts {
		subtopicPart = fmt.Sprintf("/%s", strings.ReplaceAll(chanNameParts[1], ".", "/"))
	}

	reqURL := fmt.Sprintf("%s/c/%s/m%s", sdk.httpAdapterURL, chanID, subtopicPart)

	_, _, err := sdk.processRequest(ctx, http.MethodPost, reqURL, ClientPrefix+key, []byte(msg), nil, http.StatusAccepted)

	return err
}

func (sdk *mgSDK) SetContentType(ct ContentType) errors.SDKError {
	if ct != CTJSON && ct != CTJSONSenML && ct != CTBinary {
		return errors.NewSDKError(apiutil.ErrUnsupportedContentType)
	}

	sdk.msgContentType = ct

	return nil
}
