package connectors

import (
	"net/http"

	"github.com/dexidp/dex/pkg/webhook/config"
	"github.com/dexidp/dex/pkg/webhook/helpers"
	"github.com/dexidp/dex/storage"
)

type FilterCaller interface {
	CallHook(connectors []storage.Connector, r *http.Request) ([]storage.Connector, error)
}

type ConnectorFilterHook struct {
	// Name is the name of the webhook
	Name string `json:"name"`
	// Config is the configuration of the webhook
	FilterInvoker FilterCaller `json:"filterInvoker"`
}

var _ FilterCaller = &WebhookFilterCaller{}

type WebhookFilterCaller struct {
	RequestScope    *config.HookRequestScope
	transportHelper helpers.WebhookHTTPHelpers
}

type ConnectorContext struct {
	// ID that will uniquely identify the connector object.
	ID string `json:"id"`
	// The Type of the connector. E.g. 'oidc' or 'ldap'
	Type string `json:"type"`
	// The Name of the connector that is used when displaying it to the end user.
	Name string `json:"name"`
}

type RequestContext struct {
	// Headers is the headers of the request
	Headers map[string][]string `json:"headers"`
	// Params is the params of the request
	Params map[string][]string `json:"params"`
}

type FilterWebhookPayload struct {
	Connectors []ConnectorContext `json:"connID"`
	Request    RequestContext     `json:"requestContext"`
}
