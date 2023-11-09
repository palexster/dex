//go:generate go run -mod mod go.uber.org/mock/mockgen -destination=./mocks/mock_caller.go -package=connectors --source=types.go FilterCaller
package connectors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dexidp/dex/pkg/webhook/config"
	"github.com/dexidp/dex/pkg/webhook/helpers"
	"github.com/dexidp/dex/storage"
)

func NewConnectorFilter(hook *config.ConnectorFilterHook) (*ConnectorFilterHook, error) {
	var hookInvoker FilterCaller
	switch hook.Type {
	case config.External:
		h, err := helpers.NewWebhookHTTPHelpers(hook.Config)
		if err != nil {
			return nil, fmt.Errorf("could not create webhook http helpers: %w", err)
		}
		hookInvoker = NewFilterCaller(h, hook.RequestScope)
	default:
		return nil, fmt.Errorf("unknown type: %s", hook.Type)
	}
	return &ConnectorFilterHook{
		Name:          hook.Name,
		FilterInvoker: hookInvoker,
	}, nil
}

func (f WebhookFilterCaller) callHook(connectors []ConnectorContext, req RequestContext) ([]ConnectorContext, error) {
	toMarshal := FilterWebhookPayload{
		Connectors: connectors,
		Request:    req,
	}

	payload, err := json.Marshal(toMarshal)
	if err != nil {
		return nil, fmt.Errorf("could not serialize claims: %w", err)
	}

	body, err := f.transportHelper.CallWebhook(payload)
	if err != nil {
		return nil, fmt.Errorf("could not call webhook: %w", err)
	}
	var res []ConnectorContext

	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("could not unmarshal response: %w", err)
	}

	return res, nil
}

func NewFilterCaller(h helpers.WebhookHTTPHelpers,
	context *config.HookRequestScope,
) *WebhookFilterCaller {
	return &WebhookFilterCaller{
		RequestScope:    context,
		transportHelper: h,
	}
}

func (f WebhookFilterCaller) CallHook(connectors []storage.Connector,
	r *http.Request,
) ([]storage.Connector, error) {
	payload := createConnectorWebhookPayload(f.RequestScope, connectors, r)
	filteredConnectors, err := f.callHook(payload.Connectors, payload.Request)
	if err != nil {
		return nil, err
	}
	return unwrapConnectorWebhookPayload(filteredConnectors, connectors), nil
}
