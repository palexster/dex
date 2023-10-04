package webhook

import (
	"context"
	"github.com/mesosphere/kommander/v2/clientapis/pkg/apis/workspaces/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetWsNamespaceMap(cl client.Client) (map[string]string, error) {
	workspaceMap := make(map[string]string)
	wsList := &v1alpha1.WorkspaceList{}
	err := cl.List(context.Background(), wsList)
	if err != nil {
		return nil, err
	}

	for _, ws := range wsList.Items {
		if ws.Status.NamespaceRef == nil || ws.Status.NamespaceRef.Name == "" {
			continue
		}
		workspaceMap[ws.Status.NamespaceRef.Name] = ws.Name
	}
	return workspaceMap, nil
}
