package webhook

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/mesosphere/kommander/v2/clientapis/pkg/apis/workspaces/v1alpha1"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/dexidp/dex/storage"
)

func forgeWebhook() ConnectorWebhookFilter {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	initObjs := []client.Object{
		&v1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: v1alpha1.WorkspaceSpec{
				ClusterLabels: nil,
				NamespaceName: ptr.String("test"),
			},
			Status: v1alpha1.WorkspaceStatus{
				NamespaceRef: &v12.LocalObjectReference{
					Name: "test",
				},
			},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(initObjs...).Build()
	return &ConnectorWebhookImpl{
		cl: cl,
	}
}

func Test_Filter_Pre_Tenant_Filter(t *testing.T) {
	webhook := forgeWebhook()
	connectors := []storage.Connector{
		{
			ID:     "test",
			Type:   "local",
			Name:   "test",
			Config: nil,
		},
	}
	r := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: "/dex/auth",
			RawQuery: "client_id=example-app&redirect_uri=http%3A%2F%2F127.0.0." +
				"1%3A5556%2Fcallback&scope=openid&response_type=code&response_mode=fragment&state=i2qfnocqvhb&nonce" +
				"=4uwvq4p3rts&tenant-id=test",
		},
		Proto: "HTTP/1.1",
		Host:  "127.0.0.1:5556",
		Form: url.Values{
			"client_id":     []string{"example-app"},
			"redirect_uri":  []string{"http://127.0.0.1:5556/callback"},
			"scope":         []string{"openid"},
			"response_type": []string{"code"},
			"response_mode": []string{"fragment"},
			"state":         []string{"i2qfnocqvhb"},
			"nonce":         []string{"4uwvq4p3rts"},
			"tenant-id":     []string{"test"},
		},
		RemoteAddr: "127.0.0.1:48126",
		RequestURI: "/dex/auth?client_id=example-app&redirect_uri=http%3A%2F%2F127.0.0." +
			"1%3A5556%2Fcallback&scope=openid&response_type=code&response_mode=fragment&state=i2qfnocqvhb&nonce" +
			"=4uwvq4p3rts&tenant-id=test",
	}
	filteredConnectors, err := webhook.FilterConnectors(connectors, r)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(filteredConnectors))
}

func Test_Filter_Tenant_Filter(t *testing.T) {
	webhook := forgeWebhook()
	connectors := []storage.Connector{
		{
			ID:   "mock_test_1",
			Type: "mockCallback",
			Name: "test",
		},
		{
			ID:   "mock_test_2",
			Type: "mockCallback",
			Name: "test",
		},
		{
			ID: "preexistent-2",
		},
	}
	r := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: "/dex/auth",
			RawQuery: "client_id=example-app&redirect_uri=http%3A%2F%2F127.0.0." +
				"1%3A5556%2Fcallback&scope=openid&response_type=code&response_mode=fragment&state=i2qfnocqvhb&nonce" +
				"=4uwvq4p3rts&tenant-id=test",
			Fragment:    "",
			RawFragment: "",
		},
		Proto: "HTTP/1.1",
		Host:  "127.0.0.1:5556",
		Form: url.Values{
			"client_id":     []string{"example-app"},
			"redirect_uri":  []string{"http://127.0.0.1:5556/callback"},
			"scope":         []string{"openid"},
			"response_type": []string{"code"},
			"response_mode": []string{"fragment"},
			"state":         []string{"i2qfnocqvhb"},
			"nonce":         []string{"4uwvq4p3rts"},
			"tenant-id":     []string{"test"},
		},
		RemoteAddr: "127.0.0.1:48126",
		RequestURI: "/dex/auth?client_id=example-app&redirect_uri=http%3A%2F%2F127.0.0." +
			"1%3A5556%2Fcallback&scope=openid&response_type=code&response_mode=fragment&state=i2qfnocqvhb&nonce" +
			"=4uwvq4p3rts&tenant-id=test",
	}
	filteredConnectors, err := webhook.FilterConnectors(connectors, r)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(filteredConnectors))
	assert.Equal(t, "mock_test_1", filteredConnectors[0].ID)
	assert.Equal(t, "mock_test_2", filteredConnectors[1].ID)
	assert.Equal(t, "preexistent-2", filteredConnectors[2].ID)
}

func Test_Filter_Different_Tenant_Filter(t *testing.T) {
	webhook := forgeWebhook()
	connectors := []storage.Connector{
		{
			ID:     "mock_test_2",
			Type:   "mockCallback",
			Name:   "test",
			Config: nil,
		},
		{
			ID:     "mock_othertenant_1",
			Type:   "mockCallback",
			Name:   "test",
			Config: nil,
		},
	}
	r := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme:     "",
			Opaque:     "",
			User:       nil,
			Host:       "",
			Path:       "/dex/auth",
			RawPath:    "",
			OmitHost:   false,
			ForceQuery: false,
			RawQuery: "client_id=example-app&redirect_uri=http%3A%2F%2F127.0.0." +
				"1%3A5556%2Fcallback&scope=openid&response_type=code&response_mode=fragment&state=i2qfnocqvhb&nonce" +
				"=4uwvq4p3rts&tenant-id=test",
			Fragment:    "",
			RawFragment: "",
		},
		Proto: "HTTP/1.1",
		Host:  "127.0.0.1:5556",
		Form: url.Values{
			"client_id":     []string{"example-app"},
			"redirect_uri":  []string{"http://127.0.0.1:5556/callback"},
			"scope":         []string{"openid"},
			"response_type": []string{"code"},
			"response_mode": []string{"fragment"},
			"state":         []string{"i2qfnocqvhb"},
			"nonce":         []string{"4uwvq4p3rts"},
			"tenant-id":     []string{"test"},
		},
		RemoteAddr: "127.0.0.1:48126",
		RequestURI: "/dex/auth?client_id=example-app&redirect_uri=http%3A%2F%2F127.0.0." +
			"1%3A5556%2Fcallback&scope=openid&response_type=code&response_mode=fragment&state=i2qfnocqvhb&nonce" +
			"=4uwvq4p3rts&tenant-id=test",
	}
	filteredConnectors, err := webhook.FilterConnectors(connectors, r)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(filteredConnectors))
	assert.Equal(t, "mock_test_2", filteredConnectors[0].ID)
}

func Test_Filter_No_Tenant_Parameter(t *testing.T) {
	webhook := forgeWebhook()
	connectors := []storage.Connector{
		{
			ID:     "mock_test_2",
			Type:   "mockCallback",
			Name:   "test",
			Config: nil,
		},
		{
			ID:     "mock_othertenant_1",
			Type:   "mockCallback",
			Name:   "test",
			Config: nil,
		},
		{
			ID:              "preexistent-2",
			Type:            "",
			Name:            "",
			ResourceVersion: "",
			Config:          nil,
		},
	}
	r := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme:     "",
			Opaque:     "",
			User:       nil,
			Host:       "",
			Path:       "/dex/auth",
			RawPath:    "",
			OmitHost:   false,
			ForceQuery: false,
			RawQuery: "client_id=example-app&redirect_uri=http%3A%2F%2F127.0.0." +
				"1%3A5556%2Fcallback&scope=openid&response_type=code&response_mode=fragment&state=i2qfnocqvhb&nonce" +
				"=4uwvq4p3rts",
			Fragment:    "",
			RawFragment: "",
		},
		Proto: "HTTP/1.1",
		Host:  "127.0.0.1:5556",
		Form: url.Values{
			"client_id":     []string{"example-app"},
			"redirect_uri":  []string{"http://127.0.0.1:5556/callback"},
			"scope":         []string{"openid"},
			"response_type": []string{"code"},
			"response_mode": []string{"fragment"},
			"state":         []string{"i2qfnocqvhb"},
			"nonce":         []string{"4uwvq4p3rts"},
		},
		RemoteAddr: "127.0.0.1:48126",
		RequestURI: "/dex/auth?client_id=example-app&redirect_uri=http%3A%2F%2F127.0.0." +
			"1%3A5556%2Fcallback&scope=openid&response_type=code&response_mode=fragment&state=i2qfnocqvhb&nonce" +
			"=4uwvq4p3rts",
	}
	filteredConnectors, err := webhook.FilterConnectors(connectors, r)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(filteredConnectors))
	assert.Equal(t, "preexistent-2", filteredConnectors[0].ID)
}
