package webhook

import (
	"context"
	"net/http"
	"strings"

	"github.com/mesosphere/kommander/v2/clientapis/pkg/apis/workspaces/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/dexidp/dex/storage"
)

// ConnectorWebhookFilter is an interface for filtering connectors
type ConnectorWebhookFilter interface {
	FilterConnectors(connectors []storage.Connector, r *http.Request) ([]storage.Connector, error)
}

// ConnectorWebhookImpl is an implementation of ConnectorWebhookFilter
type ConnectorWebhookImpl struct {
	cl client.Client
}

var _ ConnectorWebhookFilter = &ConnectorWebhookImpl{}

// NewConnectorWebhookFilter returns a new instance of connectorWebhookImpl
func NewConnectorWebhookFilter(kubeconfigpath string) (*ConnectorWebhookImpl, error) {
	cl, err := initializeClient(kubeconfigpath)
	if err != nil {
		return nil, err
	}

	return &ConnectorWebhookImpl{
		cl: cl,
	}, nil
}

func initializeClient(kubeconfigpath string) (client.Client, error) {
	var cfg *rest.Config
	var err error
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
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

func retrieveWorkspace(cl client.Client, workspaceName string) (*v1alpha1.Workspace, error) {
	// Get specific workspace
	w := &v1alpha1.Workspace{}
	err := cl.Get(context.Background(), client.ObjectKey{Name: workspaceName}, w)
	if client.IgnoreNotFound(err) != nil {
		return nil, err
	}
	return w, nil
}

func (c *ConnectorWebhookImpl) FilterConnectors(connectors []storage.Connector, r *http.Request) ([]storage.Connector, error) {
	// Parse query parameters
	queryParams := r.URL.Query()

	// Get the specific query parameter "key"
	key := queryParams.Get("tenant-id")
	if key != "" {
		ws, err := retrieveWorkspace(c.cl, key)
		if err != nil {
			return nil, err
		}
		if ws != nil && ws.Status.NamespaceRef != nil {
			key = ws.Status.NamespaceRef.Name
		} else {
			key = ""
		}
	}

	return filterConnectors(connectors, key), nil
}

func retrieveNamespaceFromConnectorID(connector storage.Connector) string {
	if len(strings.Split(connector.ID, "_")) < 3 {
		return ""
	}
	return strings.Split(connector.ID, "_")[1]
}

func filterConnectors(connectors []storage.Connector, key string) []storage.Connector {
	var filteredConnectors []storage.Connector
	for _, c := range connectors {
		if c.Type == "local" {
			filteredConnectors = append(filteredConnectors, c)
		} else if retrieveNamespaceFromConnectorID(c) == key {
			filteredConnectors = append(filteredConnectors, c)
		} else if retrieveNamespaceFromConnectorID(c) == "" {
			filteredConnectors = append(filteredConnectors, c)
		}
	}
	return filteredConnectors
}
