package connectors

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dexidp/dex/pkg/webhook/config"
	"github.com/dexidp/dex/storage"
)

func Test_CreatePayloadEmpty(t *testing.T) {
	payload := createConnectorWebhookPayload(&config.HookRequestScope{}, []storage.Connector{}, &http.Request{})
	assert.Equal(t, payload, &FilterWebhookPayload{
		Connectors: []ConnectorContext{},
		Request: RequestContext{
			Headers: map[string][]string{},
			Params:  map[string][]string{},
		},
	})
}

func Test_CreatePayload_ScopeValidation(t *testing.T) {
	payload := createConnectorWebhookPayload(&config.HookRequestScope{
		Headers: []string{"header1", "header2"},
		Params:  []string{"param1", "param2"},
	}, []storage.Connector{}, &http.Request{
		Header: map[string][]string{
			"header1": {"value1"},
			"header2": {"value2"},
			"header3": {"value3"},
		},
		URL: &url.URL{
			RawQuery: "param1=value1&param2=value2&param3=value3",
		},
	})
	assert.Equal(t, payload, &FilterWebhookPayload{
		Connectors: []ConnectorContext{},
		Request: RequestContext{
			Headers: map[string][]string{
				"header1": {"value1"},
				"header2": {"value2"},
			},
			Params: map[string][]string{
				"param1": {"value1"},
				"param2": {"value2"},
			},
		},
	})
}

func Test_CreatePayload_ScopeConnectorsValidation(t *testing.T) {
	payload := createConnectorWebhookPayload(&config.HookRequestScope{}, []storage.Connector{
		{ID: "test1", Type: "ldap", Name: "test1", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
	}, &http.Request{})
	assert.Equal(t, payload, &FilterWebhookPayload{
		Request: RequestContext{
			Headers: map[string][]string{},
			Params:  map[string][]string{},
		},
		Connectors: []ConnectorContext{
			{ID: "test1", Type: "ldap", Name: "test1"},
		},
	})
}

func Test_MultipleConnectorValidation(t *testing.T) {
	payload := createConnectorWebhookPayload(&config.HookRequestScope{}, []storage.Connector{
		{ID: "test1", Type: "ldap", Name: "test1", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
		{ID: "test2", Type: "ldap", Name: "test2", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
	}, &http.Request{})
	assert.Equal(t, payload, &FilterWebhookPayload{
		Request: RequestContext{
			Headers: map[string][]string{},
			Params:  map[string][]string{},
		},
		Connectors: []ConnectorContext{
			{ID: "test1", Type: "ldap", Name: "test1"},
			{ID: "test2", Type: "ldap", Name: "test2"},
		},
	})
}

func Test_UnwrapConnectorWebhookPayload_Empty(t *testing.T) {
	filteredConnectors := []ConnectorContext{}
	initialConnectors := []storage.Connector{}
	mappedConnectors := unwrapConnectorWebhookPayload(filteredConnectors, initialConnectors)
	assert.Equal(t, mappedConnectors, []storage.Connector{})
}

func Test_UnwrapConnectorWebhookPayload_LocalLDAP(t *testing.T) {
	filteredConnectors := []ConnectorContext{
		{ID: "test2", Type: "ldap", Name: "test2"},
		{ID: "test3", Type: "ldap", Name: "test3"},
	}
	initialConnectors := []storage.Connector{
		{ID: "test1", Type: "ldap", Name: "test1", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
		{ID: "test2", Type: "ldap", Name: "test2", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
		{ID: "test3", Type: "ldap", Name: "test3", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
	}
	mappedConnectors := unwrapConnectorWebhookPayload(filteredConnectors, initialConnectors)
	assert.Equal(t, mappedConnectors, []storage.Connector{
		{ID: "test2", Type: "ldap", Name: "test2", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
		{ID: "test3", Type: "ldap", Name: "test3", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
	})
}

func Test_UnwrapConnectorWebhookPayload_DifferentOrder(t *testing.T) {
	filteredConnectors := []ConnectorContext{
		{ID: "test3", Type: "ldap", Name: "test3"},
		{ID: "test2", Type: "ldap", Name: "test2"},
	}
	initialConnectors := []storage.Connector{
		{ID: "test1", Type: "ldap", Name: "test1", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
		{ID: "test2", Type: "ldap", Name: "test2", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
		{ID: "test3", Type: "ldap", Name: "test3", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
	}
	mappedConnectors := unwrapConnectorWebhookPayload(filteredConnectors, initialConnectors)
	assert.Equal(t, mappedConnectors, []storage.Connector{
		{ID: "test3", Type: "ldap", Name: "test3", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
		{ID: "test2", Type: "ldap", Name: "test2", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
	})
}

func Test_UnwrapConnectorWebhookPayload_LocalOIDC(t *testing.T) {
	filteredConnectors := []ConnectorContext{
		{ID: "test3", Type: "oidc", Name: "test3"},
	}
	initialConnectors := []storage.Connector{
		{ID: "test1", Type: "oidc", Name: "test1", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
		{ID: "test2", Type: "oidc", Name: "test2", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
		{ID: "test3", Type: "oidc", Name: "test3", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
	}
	mappedConnectors := unwrapConnectorWebhookPayload(filteredConnectors, initialConnectors)
	assert.Equal(t, mappedConnectors, []storage.Connector{
		{ID: "test3", Type: "oidc", Name: "test3", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
	})
}

func Test_UnwrapConnectorWebhookPayload_NotExisting(t *testing.T) {
	filteredConnectors := []ConnectorContext{
		{ID: "test4", Type: "oidc", Name: "test4"},
	}
	initialConnectors := []storage.Connector{
		{ID: "test1", Type: "oidc", Name: "test1", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
		{ID: "test2", Type: "oidc", Name: "test2", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
		{ID: "test3", Type: "oidc", Name: "test3", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
	}
	mappedConnectors := unwrapConnectorWebhookPayload(filteredConnectors, initialConnectors)
	assert.Equal(t, mappedConnectors, []storage.Connector{})
}

func Test_UnwrapConnectorWebhookPayload_NotExistingTest(t *testing.T) {
	filteredConnectors := []ConnectorContext{
		{ID: "test4", Type: "oidc", Name: "test4"},
		{ID: "test3", Type: "oidc", Name: "test3"},
	}
	initialConnectors := []storage.Connector{
		{ID: "test1", Type: "oidc", Name: "test1", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
		{ID: "test2", Type: "oidc", Name: "test2", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
		{ID: "test3", Type: "oidc", Name: "test3", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
	}
	mappedConnectors := unwrapConnectorWebhookPayload(filteredConnectors, initialConnectors)
	assert.Equal(t, mappedConnectors, []storage.Connector{
		{ID: "test3", Type: "oidc", Name: "test3", ResourceVersion: "123", Config: []byte(`{"some":"data"}`)},
	})
}
