package kubernetes

import (
	"path/filepath"
	"reflect"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestContextNames(t *testing.T) {
	t.Parallel()

	config := clientcmdapi.NewConfig()
	config.Contexts["staging"] = &clientcmdapi.Context{}
	config.Contexts["operations"] = &clientcmdapi.Context{}
	config.Contexts["production"] = &clientcmdapi.Context{}

	path := filepath.Join(t.TempDir(), "config")

	if err := clientcmd.WriteToFile(*config, path); err != nil {
		t.Fatalf("write kubeconfig: %v", err)
	}

	got, err := ContextNames(path)
	if err != nil {
		t.Fatalf("ContextNames() error = %v", err)
	}

	want := []string{
		"operations",
		"production",
		"staging",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"ContextNames() = %v, want %v",
			got,
			want,
		)
	}
}
