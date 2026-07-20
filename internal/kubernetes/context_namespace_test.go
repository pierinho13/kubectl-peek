package kubernetes

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestSetContextNamespaceUpdatesCurrentContext(
	t *testing.T,
) {
	t.Parallel()

	kubeconfig := filepath.Join(t.TempDir(), "config")

	config := api.NewConfig()
	config.CurrentContext = "development"
	config.Contexts["development"] = &api.Context{
		Cluster:   "cluster",
		AuthInfo:  "user",
		Namespace: "default",
	}

	data, err := clientcmd.Write(*config)
	if err != nil {
		t.Fatalf("write kubeconfig: %v", err)
	}

	if err := os.WriteFile(kubeconfig, data, 0o600); err != nil {
		t.Fatalf("write kubeconfig file: %v", err)
	}

	contextName, err := SetContextNamespace(
		kubeconfig,
		"",
		"monitoring",
	)
	if err != nil {
		t.Fatalf("SetContextNamespace() error = %v", err)
	}

	if contextName != "development" {
		t.Fatalf(
			"expected context development, got %q",
			contextName,
		)
	}

	updated, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		t.Fatalf("load updated kubeconfig: %v", err)
	}

	got := updated.Contexts["development"].Namespace

	if got != "monitoring" {
		t.Fatalf(
			"expected namespace monitoring, got %q",
			got,
		)
	}
}
