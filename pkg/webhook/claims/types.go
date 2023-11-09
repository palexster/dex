package claims

import (
	"github.com/dexidp/dex/pkg/webhook/config"
	"github.com/dexidp/dex/pkg/webhook/helpers"
)

type MutateClaimsRequest interface {
	MutateClaims(claims map[string]interface{}, connID string) (map[string]interface{}, error)
}

type ClaimsHookCaller interface {
	CallHook(claims map[string]interface{}, connID string) (map[string]interface{}, error)
}

type ClaimsMutatingHook struct {
	// Name is the name of the webhook
	Name string `json:"name"`
	// To be modified to enum?
	Type        config.HookType `json:"type"`
	HookInvoker ClaimsHookCaller
}

var _ ClaimsHookCaller = &WebhookCallerImpl{}

type WebhookCallerImpl struct {
	AcceptedClaims  []string
	transportHelper helpers.WebhookHTTPHelpers
}
