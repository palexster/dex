package webhook

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	dexv1alpha1 "github.com/mesosphere/dex-controller/api/v1alpha1"
	"github.com/mesosphere/kommander/v2/clientapis/pkg/apis/workspaces/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/dexidp/dex/storage"
)

const (
	DefaultClientWorkspace = "kommander"
	DefaultWorkspace       = "kommander-default-workspace"
	EmptyTenantID          = ""
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
	// Get the specific query parameter "key"
	clientList, err := initializeDexClientList(c.cl)
	if err != nil {
		return nil, err
	}
	workspaceMap := make(map[string]string)
	wsList := &v1alpha1.WorkspaceList{}
	err = c.cl.List(context.Background(), wsList)
	if err != nil {
		return nil, err
	}
	for _, ws := range wsList.Items {
		if ws.Status.NamespaceRef == nil || ws.Status.NamespaceRef.Name == "" {
			continue
		}
		workspaceMap[ws.Status.NamespaceRef.Name] = ws.Name
	}
	key := parseTenantID(clientList, workspaceMap, r)
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

func initializeDexClientList(cl client.Client) (*dexv1alpha1.ClientList, error) {
	dexClientList := dexv1alpha1.ClientList{}
	err := cl.List(context.Background(), &dexClientList, client.InNamespace("kommander"))
	if err != nil {
		return nil, err
	}
	return &dexClientList, nil
}

func parseStateRequestHeader(stateParam string) (string, error) {
	// Remove the non-standard prefix, if it exists
	colonIndex := strings.Index(stateParam, ":")
	if colonIndex != -1 {
		// Remove the prefix by slicing the string
		stateParam = stateParam[colonIndex+1:]
	}

	// Parse the URL
	u, err := url.Parse(stateParam)
	if err != nil {
		return "", err
	}

	return u.Query().Get("tenant-id"), nil
}

// parseTenantID parses the tenant ID from the state request header, if not found looks to the
// query parameters "tenant-id". As last resort, it extracts the cluster workspace from the client ID, inferring
// the tenant ID from the namespace.
// Priority:
// 1. state request header
// 2. query parameter "tenant-id"
// 3. client ID
func parseTenantID(dexClientList *dexv1alpha1.ClientList, workspaceMap map[string]string, r *http.Request) string {
	// If the state request header contains the tenant ID, we use it.
	receivedState := r.URL.Query().Get("state")
	if receivedState != "" && strings.LastIndex(receivedState, "tenant-id=") != -1 {
		tenantID, err := parseStateRequestHeader(receivedState)
		if err == nil {
			return tenantID
		}
	}

	// If the query parameter "tenant-id" is present, we use it.
	if r.URL.Query().Get("tenant-id") != EmptyTenantID {
		return r.URL.Query().Get("tenant-id")
	}

	// removing the client ID prefix dex-controller-
	reqClientID := strings.TrimPrefix(r.URL.Query().Get("client_id"), "dex-controller-")
	// Otherwise, we use the "client_id" parameter.
	if reqClientID == "" || len(dexClientList.Items) == 0 {
		return EmptyTenantID
	}

	for _, c := range dexClientList.Items {
		if c.Name == reqClientID {
			ns := strings.Split(c.Spec.DisplayName, "/")[0]
			if ns == DefaultClientWorkspace {
				return EmptyTenantID
			}
			return workspaceMap[ns]
		}
	}
	return EmptyTenantID
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
		switch {
		case c.Type == "local":
			filteredConnectors = append(filteredConnectors, c)
		// If the connector is present in the kommander defuault namespace,
		// we always want to return it since it belongs to the admin organization.
		case retrieveNamespaceFromConnectorID(c) == DefaultWorkspace:
			filteredConnectors = append(filteredConnectors, c)
		case retrieveNamespaceFromConnectorID(c) == key:
			filteredConnectors = append(filteredConnectors, c)
		case retrieveNamespaceFromConnectorID(c) == "":
			filteredConnectors = append(filteredConnectors, c)
		}
	}
	return filteredConnectors
}
