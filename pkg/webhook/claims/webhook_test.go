package claims

import (
	"github.com/aws/smithy-go/ptr"
	dexv1alpha1 "github.com/mesosphere/dex-controller/api/v1alpha1"
	"github.com/mesosphere/kommander/v2/clientapis/pkg/apis/workspaces/v1alpha1"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func forgeWebhook() IDTokenMutatingFilter {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	_ = dexv1alpha1.AddToScheme(scheme)
	initObjs := []client.Object{
		&v1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "tenant1-4jw2v",
				Labels: map[string]string{},
			},
			Spec: v1alpha1.WorkspaceSpec{
				ClusterLabels: nil,
				NamespaceName: ptr.String("tenant1-4jw2v-4zsmq"),
			},
			Status: v1alpha1.WorkspaceStatus{
				NamespaceRef: &v12.LocalObjectReference{
					Name: "tenant1-4jw2v-4zsmq",
				},
			},
		},
		&v1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "not-enabled",
			},
			Spec: v1alpha1.WorkspaceSpec{
				ClusterLabels: nil,
			},
			Status: v1alpha1.WorkspaceStatus{
				NamespaceRef: &v12.LocalObjectReference{
					Name: "not-enabled",
				},
			},
		},
		&dexv1alpha1.Client{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "implicit-app",
				Namespace: "kommander",
			},
			Spec: dexv1alpha1.ClientSpec{
				DisplayName: "test/cluster1",
			},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(initObjs...).Build()
	return &IDTokenMutatingWebhookImpl{
		cl: cl,
	}
}

func Test_WorkspaceNotFound(t *testing.T) {
	s := forgeWebhook()
	p, err := s.MutateClaims(&IdTokenClaimsPayload{
		Groups:            []string{"group1", "group2"},
		PreferredUsername: "user",
	}, "mock_not-existing_mock")
	assert.Error(t, err, "Workspace not found")
	assert.Nil(t, p)
}

func Test_NotValidConnector(t *testing.T) {
	s := forgeWebhook()
	p, err := s.MutateClaims(&IdTokenClaimsPayload{
		Groups:            []string{"group1", "group2"},
		PreferredUsername: "user",
	}, "not-valid-connector")
	assert.Error(t, err, "Workspace not found")
	assert.Nil(t, p)
}

func Test_TenantedConnector(t *testing.T) {
	s := forgeWebhook()
	p, err := s.MutateClaims(&IdTokenClaimsPayload{
		Groups:            []string{"group1", "group2"},
		PreferredUsername: "user",
	}, "mock_tenant1-4jw2v-4zsmq_mock")
	assert.Nil(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "tenant1-4jw2v:user", p.PreferredUsername)
	assert.Equal(t, 2, len(p.Groups))
	assert.Equal(t, "tenant1-4jw2v:group1", p.Groups[0])
	assert.Equal(t, "tenant1-4jw2v:group2", p.Groups[1])
}

func Test_GlobalConnector(t *testing.T) {
	s := forgeWebhook()
	p, err := s.MutateClaims(&IdTokenClaimsPayload{
		Groups:            []string{"group1", "group2"},
		PreferredUsername: "user",
	}, "mock_kommander_mock")
	assert.Nil(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "user", p.PreferredUsername)
	assert.Equal(t, 2, len(p.Groups))
	assert.Equal(t, "group1", p.Groups[0])
	assert.Equal(t, "group2", p.Groups[1])
}
