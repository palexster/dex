package claims

import (
	"errors"
	"github.com/dexidp/dex/pkg/webhook"
	dexv1alpha1 "github.com/mesosphere/dex-controller/api/v1alpha1"
	"github.com/mesosphere/kommander/v2/clientapis/pkg/apis/workspaces/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"strings"
)

// ConnectorWebhookFilter is an interface for filtering connectors
type IDTokenMutatingFilter interface {
	MutateClaims(claims *IdTokenClaimsPayload, connID string) (*IdTokenClaimsPayload, error)
}

type IdTokenClaimsPayload struct {
	Groups            []string `json:"groups"`
	PreferredUsername string
}

// ConnectorWebhookImpl is an implementation of ConnectorWebhookFilter
type IDTokenMutatingWebhookImpl struct {
	cl client.Client
}

var _ IDTokenMutatingFilter = &IDTokenMutatingWebhookImpl{}

func initializeClient(kubeconfigpath string) (client.Client, error) {
	var cfg *rest.Config
	var err error
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	_ = dexv1alpha1.AddToScheme(scheme)
	if kubeconfigpath == "" {
		cfg = config.GetConfigOrDie()
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfigpath)
		if err != nil {
			return nil, err
		}
	}

	return client.New(cfg, client.Options{Scheme: scheme})
}

// NewConnectorWebhookFilter returns a new instance of connectorWebhookImpl
func NewIDTokenMutatingWebhookFilter(kubeconfigpath string) (*IDTokenMutatingWebhookImpl, error) {
	cl, err := initializeClient(kubeconfigpath)
	if err != nil {
		return nil, err
	}

	return &IDTokenMutatingWebhookImpl{
		cl: cl,
	}, nil
}

func (s IDTokenMutatingWebhookImpl) MutateClaims(claims *IdTokenClaimsPayload, connID string) (
	*IdTokenClaimsPayload, error) {
	idParsing := strings.Split(connID, "_")
	if len(idParsing) != 3 {
		return nil, errors.New("Invalid connector ID name")
	}
	workspaceNamespace := idParsing[1]

	// If the workspace is kommander, we don't need to do anything
	// this is a cluster-wide connector not associated with any tenant
	if workspaceNamespace == "kommander" {
		return claims, nil
	}

	namespaceMap, err := webhook.GetWsNamespaceMap(s.cl)
	if err != nil {
		return nil, err
	}

	workspaceName, ok := namespaceMap[workspaceNamespace]
	if !ok {
		return nil, errors.New("Workspace not found")
	}

	claims.Groups = translateGroups(workspaceName, claims.Groups)
	claims.PreferredUsername = workspaceName + ":" + claims.PreferredUsername

	return claims, nil
}

func translateGroups(prefix string, groups []string) []string {
	//	return []string{"oidc:tenant1:test"}
	if len(groups) == 0 {
		return nil
	}
	prefixedGroups := make([]string, 0)
	for _, group := range groups {
		prefixedGroups = append(prefixedGroups, prefix+":"+group)
	}
	return prefixedGroups
}
