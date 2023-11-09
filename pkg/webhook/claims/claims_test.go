package claims

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/dexidp/dex/pkg/webhook/config"
	"github.com/dexidp/dex/pkg/webhook/helpers"
)

func TestNewClaimsMutating(t *testing.T) {
	hook, err := NewClaimMutatingHook(&config.ClaimsMutatingHook{
		Name:           "test",
		Type:           config.External,
		AcceptedClaims: []string{"claim1", "claim2"},
		Config: &config.WebhookConfig{
			URL:                "https://test.com",
			InsecureSkipVerify: true,
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, hook)
	assert.Equal(t, hook.Name, "test")
	assert.Equal(t, hook.Type, config.External)
	assert.IsType(t, hook.HookInvoker, &WebhookCallerImpl{})
}

func TestNewClaimsMutating_UnknownType(t *testing.T) {
	hook, err := NewClaimMutatingHook(&config.ClaimsMutatingHook{
		Name:           "test",
		Type:           "Unknown",
		AcceptedClaims: []string{"claim1", "claim2"},
		Config: &config.WebhookConfig{
			URL:                "https://test.com",
			InsecureSkipVerify: true,
		},
	})
	assert.Error(t, err)
	assert.Nil(t, hook)
}

func TestNewWebhookCaller(t *testing.T) {
	h, err := helpers.NewWebhookHTTPHelpers(&config.WebhookConfig{
		URL:                "https://test.com",
		InsecureSkipVerify: true,
	})
	assert.NoError(t, err)
	d := NewWebhookCaller(h, []string{"claim1", "claim2"})
	assert.NotNil(t, d)
	assert.Equal(t, d.AcceptedClaims, []string{"claim1", "claim2"})
	assert.IsType(t, d.transportHelper, h)
}

func TestCallHook_Logic_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	h := helpers.NewMockWebhookHTTPHelpers(ctrl)
	h.EXPECT().CallWebhook(gomock.Any()).Return(nil, assert.AnError)
	d := NewWebhookCaller(h, []string{"claim1", "claim2"})
	hook, err := d.CallHook(map[string]interface{}{"claim1": "value1", "claim2": "value2"}, "test")
	assert.Error(t, err)
	assert.Nil(t, hook)
}

func TestCallHook_Logic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	h := helpers.NewMockWebhookHTTPHelpers(ctrl)
	h.EXPECT().CallWebhook([]byte(`{"connID":"test","claims":{"claim1":"value1"}}`)).Return([]byte(
		`{"connID": "test", "claims": { "claim1" : "value1" } }`), nil)
	d := NewWebhookCaller(h, []string{"claim1", "claim3"})
	hook, err := d.CallHook(map[string]interface{}{"claim1": "value1", "claim2": "value2"}, "test")
	assert.NoError(t, err)
	assert.NotNil(t, hook)
}
