package connectors

import (
	"net/http"

	"golang.org/x/exp/slices"

	"github.com/dexidp/dex/pkg/webhook/config"
	"github.com/dexidp/dex/storage"
)

func createConnectorWebhookPayload(requestScope *config.HookRequestScope, connectors []storage.Connector,
	r *http.Request,
) *FilterWebhookPayload {
	payload := &FilterWebhookPayload{
		Connectors: []ConnectorContext{},
		Request:    RequestContext{},
	}
	for _, c := range connectors {
		payload.Connectors = append(payload.Connectors, ConnectorContext{
			ID:   c.ID,
			Type: c.Type,
			Name: c.Name,
		})
	}
	payload.Request.Params = make(map[string][]string)
	if r != nil && r.URL != nil {
		for k, v := range r.URL.Query() {
			if slices.Contains(requestScope.Params, k) {
				payload.Request.Params[k] = v
			}
		}
	}
	payload.Request.Headers = make(map[string][]string)
	for k, v := range r.Header {
		if slices.Contains(requestScope.Headers, k) {
			payload.Request.Headers[k] = v
		}
	}
	return payload
}

func unwrapConnectorWebhookPayload(filteredConnectors []ConnectorContext,
	initialConnectors []storage.Connector,
) []storage.Connector {
	mappedConnectors := make([]storage.Connector, 0)
	for _, filteredConnector := range filteredConnectors {
		for _, initialConnector := range initialConnectors {
			if filteredConnector.ID == initialConnector.ID {
				mappedConnectors = append(mappedConnectors, initialConnector)
			}
		}
	}
	return mappedConnectors
}
