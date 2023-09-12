package webhook

import (
	dexv1alpha1 "github.com/mesosphere/dex-controller/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/url"
	"testing"
)

func TestParse_ParamTenantID(t *testing.T) {
	r := &http.Request{
		URL: &url.URL{
			Scheme:      "",
			Opaque:      "",
			User:        nil,
			Host:        "",
			Path:        "/dex/auth",
			RawPath:     "",
			OmitHost:    false,
			ForceQuery:  false,
			RawQuery:    "tenant-id=tenant-1-685lr&client_id=dex-controller-dextfa-client-host-cluster-dcpl6&redirect_uri=https%3A%2F%2Fexample.com%2F_oauth&response_type=code&scope=openid+profile+email+groups&state=f987a4f5b0dfb88870ac3c98e05e5b66%3Ahttps%3A%2F%2Fexample.com%2Fdkp%2Fkommander%2Fdashboard%2F",
			Fragment:    "",
			RawFragment: "",
		},
		Form: url.Values{
			"response_mode": []string{"fragment"},
			"nonce":         []string{"4uwvq4p3rts"},
			"tenant-id":     []string{"tenant-1-685lr"},
			"client_id":     []string{"dex-controller-dextfa-client-host-cluster-dcpl6"},
			"redirect_uri": []string{"https://example." +
				"com/_oauth"},
			"response_type": []string{"code"},
			"scope":         []string{"openid profile email groups"},
			"state": []string{"f987a4f5b0dfb88870ac3c98e05e5b66:https://af31a3bd727a943bd902889e6f964f39" +
				"-399386962.us-west-2.elb.amazonaws.com/dkp/kommander/dashboard/"},
		},
		RemoteAddr: "127.0.0.1:48126",
		RequestURI: "dex/auth?tenant-id=tenant-1-685lr&client_id=dex-controller-dextfa-client-host-cluster-dcpl6&redirect_uri=https%3A%2F%2Fexample.com%2F_oauth&response_type=code&scope=openid+profile+email+groups&state=f987a4f5b0dfb88870ac3c98e05e5b66%3Ahttps%3A%2F%2Fexample.com%2Fdkp%2Fkommander%2Fdashboard%2F",
	}
	tenantID := parseTenantID(nil, nil, r)
	assert.Equal(t, "tenant-1-685lr", tenantID)
}

func TestParse_StateTenantID(t *testing.T) {
	// 'redirect_uri':	https://example.com/_oauth
	// 'response_type':	code
	// 'scope':	openid profile email groups
	// 'state':	f987a4f5b0dfb88870ac3c98e05e5b66:https://example.com/dkp/kommander/dashboard/?tenant-id=tenant-1-685lr
	r := &http.Request{
		URL: &url.URL{
			Scheme:      "",
			Opaque:      "",
			User:        nil,
			Host:        "",
			Path:        "/dex/auth",
			RawPath:     "",
			OmitHost:    false,
			ForceQuery:  false,
			RawQuery:    "client_id=dex-controller-dextfa-client-host-cluster-dcpl6&redirect_uri=https%3A%2F%2Fexample.com%2F_oauth&response_type=code&scope=openid+profile+email+groups&state=f987a4f5b0dfb88870ac3c98e05e5b66%3Ahttps%3A%2F%2Fexample.com%2Fdkp%2Fkommander%2Fdashboard%2F%3Ftenant-id%3Dtenant-1-685lr",
			Fragment:    "",
			RawFragment: "",
		},
		Form: url.Values{
			"response_mode": []string{"fragment"},
			"nonce":         []string{"4uwvq4p3rts"},
			"client_id":     []string{"dex-controller-dextfa-client-host-cluster-dcpl6"},
			"redirect_uri": []string{"https://example." +
				"com/_oauth"},
			"response_type": []string{"code"},
			"scope":         []string{"openid profile email groups"},
			"state":         []string{"f987a4f5b0dfb88870ac3c98e05e5b66:https://example.com/dkp/kommander/dashboard/?tenant-id=tenant-1-685lr"},
		},
		RemoteAddr: "127.0.0.1:48126",
		RequestURI: "dex/auth?client_id=dex-controller-dextfa-client-host-cluster-dcpl6&redirect_uri=https%3A%2F%2Fexample.com%2F_oauth&response_type=code&scope=openid+profile+email+groups&state=f987a4f5b0dfb88870ac3c98e05e5b66%3Ahttps%3A%2F%2Fexample.com%2Fdkp%2Fkommander%2Fdashboard%2F%3Ftenant-id%3Dtenant-1-685lr",
	}
	tenantID := parseTenantID(nil, nil, r)
	assert.Equal(t, "tenant-1-685lr", tenantID)
}

func TestParse_ClusterIDTenantID(t *testing.T) {
	// 'redirect_uri':	https://example.com/_oauth
	// 'response_type':	code
	// 'scope':	openid profile email groups
	// 'state':	f987a4f5b0dfb88870ac3c98e05e5b66:https://example.com/dkp/kommander/dashboard/?tenant-id=tenant-1-685lr
	r := &http.Request{
		URL: &url.URL{
			Scheme:      "",
			Opaque:      "",
			User:        nil,
			Host:        "",
			Path:        "/dex/auth",
			RawPath:     "",
			OmitHost:    false,
			ForceQuery:  false,
			RawQuery:    "client_id=dextfa-clientsecret-tenant2-test-wxtnm&redirect_uri=https%3A%2F%2Fexample.com%2F_oauth&response_type=code&scope=openid+profile+email+groups&state=f987a4f5b0dfb88870ac3c98e05e5b66%3Ahttps%3A%2F%2Fexample.com%2Fdkp%2Fkommander%2Fdashboard%2F",
			Fragment:    "",
			RawFragment: "",
		},
		Form: url.Values{
			"response_mode": []string{"fragment"},
			"nonce":         []string{"4uwvq4p3rts"},
			"client_id":     []string{"dextfa-clientsecret-tenant2-test-wxtnm"},
			"redirect_uri": []string{"https://example." +
				"com/_oauth"},
			"response_type": []string{"code"},
			"scope":         []string{"openid profile email groups"},
			"state":         []string{"f987a4f5b0dfb88870ac3c98e05e5b66:https://example.com/dkp/kommander/dashboard/"},
		},
		RemoteAddr: "127.0.0.1:48126",
		RequestURI: "dex/auth?client_id=dextfa-clientsecret-tenant2-test-wxtnm&redirect_uri=https%3A%2F%2Fexample.com%2F_oauth&response_type=code&scope=openid+profile+email+groups&state=f987a4f5b0dfb88870ac3c98e05e5b66%3Ahttps%3A%2F%2Fexample.com%2Fdkp%2Fkommander%2Fdashboard%2F",
	}
	dexClientList := &dexv1alpha1.ClientList{
		Items: []dexv1alpha1.Client{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dextfa-clientsecret-tenant2-test-wxtnm",
				},
				Spec: dexv1alpha1.ClientSpec{
					DisplayName: "tenant-1-99z7c-bprmg/tenant1-test",
				},
				Status: dexv1alpha1.ClientStatus{},
			},
		},
	}
	workspaceMap := map[string]string{
		"tenant-1-99z7c-bprmg": "tenant-1-685lr",
	}
	tenantID := parseTenantID(dexClientList, workspaceMap, r)
	assert.Equal(t, "tenant-1-685lr", tenantID)
}

func TestParse_NoTenantID(t *testing.T) {
	// 'redirect_uri':	https://example.com/_oauth
	// 'response_type':	code
	// 'scope':	openid profile email groups
	// 'state':	f987a4f5b0dfb88870ac3c98e05e5b66:https://example.com/dkp/kommander/dashboard/?tenant-id=tenant-1-685lr
	r := &http.Request{
		URL: &url.URL{
			Scheme:      "",
			Opaque:      "",
			User:        nil,
			Host:        "",
			Path:        "/dex/auth",
			RawPath:     "",
			OmitHost:    false,
			ForceQuery:  false,
			RawQuery:    "client_id=dextfa-clientsecret-tenant2-test-wxtnm&redirect_uri=https%3A%2F%2Fexample.com%2F_oauth&response_type=code&scope=openid+profile+email+groups&state=f987a4f5b0dfb88870ac3c98e05e5b66%3Ahttps%3A%2F%2Fexample.com%2Fdkp%2Fkommander%2Fdashboard%2F",
			Fragment:    "",
			RawFragment: "",
		},
		Form: url.Values{
			"response_mode": []string{"fragment"},
			"nonce":         []string{"4uwvq4p3rts"},
			"client_id":     []string{"dextfa-clientsecret-tenant2-test-wxtnm"},
			"redirect_uri": []string{"https://example." +
				"com/_oauth"},
			"response_type": []string{"code"},
			"scope":         []string{"openid profile email groups"},
			"state":         []string{"f987a4f5b0dfb88870ac3c98e05e5b66:https://example.com/dkp/kommander/dashboard/"},
		},
		RemoteAddr: "127.0.0.1:48126",
		RequestURI: "dex/auth?client_id=dextfa-clientsecret-tenant2-test-wxtnm&redirect_uri=https%3A%2F%2Fexample.com%2F_oauth&response_type=code&scope=openid+profile+email+groups&state=f987a4f5b0dfb88870ac3c98e05e5b66%3Ahttps%3A%2F%2Fexample.com%2Fdkp%2Fkommander%2Fdashboard%2F",
	}
	dexClientList := &dexv1alpha1.ClientList{
		Items: []dexv1alpha1.Client{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "not-existing",
				},
				Spec: dexv1alpha1.ClientSpec{
					DisplayName: "tenant-1-99z7c-bprmg/tenant1-test",
				},
				Status: dexv1alpha1.ClientStatus{},
			},
		},
	}
	workspaceMap := map[string]string{
		"tenant-1-99z7c-bprmg": "tenant-1-685lr",
	}
	tenantID := parseTenantID(dexClientList, workspaceMap, r)
	assert.Equal(t, "", tenantID)
}
