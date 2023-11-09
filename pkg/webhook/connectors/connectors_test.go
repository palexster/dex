package connectors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/dexidp/dex/pkg/webhook/config"
	"github.com/dexidp/dex/pkg/webhook/helpers"
	"github.com/dexidp/dex/storage"
)

func TestNewConnectorFilter(t *testing.T) {
	d, err := NewConnectorFilter(&config.ConnectorFilterHook{
		Name: "test",
		Type: config.External,
		RequestScope: &config.HookRequestScope{
			Headers: []string{"header1", "header2"},
			Params:  []string{"param1", "param2"},
		},
		Config: &config.WebhookConfig{
			URL:                "http://test.com",
			InsecureSkipVerify: true,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, d)
	assert.Equal(t, d.Name, "test")
	assert.IsType(t, d.FilterInvoker, &WebhookFilterCaller{})
}

func TestNewConnectorFilter_UnknownType(t *testing.T) {
	d, err := NewConnectorFilter(&config.ConnectorFilterHook{
		Name: "test",
		Type: "unknown",
		RequestScope: &config.HookRequestScope{
			Headers: []string{"header1", "header2"},
			Params:  []string{"param1", "param2"},
		},
		Config: &config.WebhookConfig{
			URL:                "http://test.com",
			InsecureSkipVerify: true,
		},
	})
	assert.Error(t, err)
	assert.Nil(t, d)
}

func TestNewFilterCaller(t *testing.T) {
	h, err := helpers.NewWebhookHTTPHelpers(&config.WebhookConfig{
		URL:                "http://test.com",
		InsecureSkipVerify: true,
	})
	assert.NoError(t, err)
	d := NewFilterCaller(h, &config.HookRequestScope{
		Headers: []string{"header1", "header2"},
		Params:  []string{"param1", "param2"},
	})
	assert.NotNil(t, d)
	assert.Equal(t, h, d.transportHelper)
	assert.Equal(t, d.RequestScope.Headers, []string{"header1", "header2"})
	assert.Equal(t, d.RequestScope.Params, []string{"param1", "param2"})
}

func TestCallHook_Logic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	h := helpers.NewMockWebhookHTTPHelpers(ctrl)
	h.EXPECT().CallWebhook(gomock.Any()).Return([]byte(`[{"id": "test", "type": "test", "name": "test"}]`), nil)
	d := NewFilterCaller(h, &config.HookRequestScope{
		Headers: []string{"header1", "header2"},
		Params:  []string{"param1", "param2"},
	})
	connectorList, err := d.CallHook([]storage.Connector{
		{
			ID:   "test",
			Type: "test",
			Name: "test",
		},
	}, &http.Request{})
	assert.NoError(t, err)
	assert.Equal(t, connectorList, []storage.Connector{
		{
			ID:   "test",
			Type: "test",
			Name: "test",
		},
	})
}

func TestCallHook_Logic_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	h := helpers.NewMockWebhookHTTPHelpers(ctrl)
	h.EXPECT().CallWebhook(gomock.Any()).Return(nil, assert.AnError)
	d := NewFilterCaller(h, &config.HookRequestScope{
		Headers: []string{"header1", "header2"},
		Params:  []string{"param1", "param2"},
	})
	connectorList, err := d.CallHook([]storage.Connector{
		{
			ID:   "test",
			Type: "test",
			Name: "test",
		},
	}, &http.Request{})
	assert.Error(t, err)
	assert.Nil(t, connectorList)
}
