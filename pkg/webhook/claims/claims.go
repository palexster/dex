package claims

import (
	"encoding/json"
	"fmt"

	"github.com/dexidp/dex/pkg/webhook/config"
	"github.com/dexidp/dex/pkg/webhook/helpers"
)

func NewClaimMutatingHook(hook *config.ClaimsMutatingHook) (*ClaimsMutatingHook, error) {
	var hookInvoker ClaimsHookCaller
	switch hook.Type {
	case config.External:
		h, err := helpers.NewWebhookHTTPHelpers(hook.Config)
		if err != nil {
			return nil, fmt.Errorf("could not create webhook http helpers: %v", err)
		}
		hookInvoker = NewWebhookCaller(h, hook.AcceptedClaims)
	default:
		return nil, fmt.Errorf("unknown type: %s", hook.Type)
	}
	return &ClaimsMutatingHook{
		Name:        hook.Name,
		Type:        hook.Type,
		HookInvoker: hookInvoker,
	}, nil
}

func NewWebhookCaller(h helpers.WebhookHTTPHelpers, acceptedClaims []string) *WebhookCallerImpl {
	return &WebhookCallerImpl{
		AcceptedClaims:  acceptedClaims,
		transportHelper: h,
	}
}

type ClaimsWebhookPayload struct {
	ConnID string                 `json:"connID"`
	Claims map[string]interface{} `json:"claims"`
}

func (w WebhookCallerImpl) callHook(claims map[string]interface{}, connID string) (map[string]interface{}, error) {
	toMarshal := ClaimsWebhookPayload{
		ConnID: connID,
		Claims: claims,
	}

	payload, err := json.Marshal(toMarshal)
	if err != nil {
		return nil, fmt.Errorf("could not serialize claims: %v", err)
	}

	body, err := w.transportHelper.CallWebhook(payload)
	if err != nil {
		return nil, fmt.Errorf("could not call webhook: %v", err)
	}
	var res map[string]interface{}

	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("could not unmarshal response: %v", err)
	}

	return res, nil
}

func (w WebhookCallerImpl) CallHook(claims map[string]interface{}, connID string) (map[string]interface{}, error) {
	filteredClaims := constrainScope(claims, w.AcceptedClaims)
	return w.callHook(filteredClaims, connID)
}
