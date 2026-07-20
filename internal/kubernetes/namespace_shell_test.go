package kubernetes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestCreateTemporaryNamespaceKubeconfig(
	t *testing.T,
) {
	t.Parallel()

	sourceDirectory := t.TempDir()
	certificatePath := filepath.Join(
		sourceDirectory,
		"ca.crt",
	)

	if err := os.WriteFile(
		certificatePath,
		[]byte("certificate-data"),
		0o600,
	); err != nil {
		t.Fatalf("write certificate: %v", err)
	}

	config := clientcmdapi.NewConfig()
	config.CurrentContext = "staging"
	config.Clusters["cluster"] = &clientcmdapi.Cluster{
		Server:               "https://example.invalid",
		CertificateAuthority: certificatePath,
	}
	config.AuthInfos["user"] = &clientcmdapi.AuthInfo{
		Token: "token",
	}
	config.Contexts["staging"] = &clientcmdapi.Context{
		Cluster:   "cluster",
		AuthInfo:  "user",
		Namespace: "default",
	}
	config.Contexts["production"] = &clientcmdapi.Context{
		Cluster:   "cluster",
		AuthInfo:  "user",
		Namespace: "production",
	}

	sourcePath := filepath.Join(
		sourceDirectory,
		"config",
	)

	if err := clientcmd.WriteToFile(
		*config,
		sourcePath,
	); err != nil {
		t.Fatalf("write source kubeconfig: %v", err)
	}

	temporaryPath, contextName, err :=
		createTemporaryNamespaceKubeconfig(
			sourcePath,
			"",
			"traefik",
			t.TempDir(),
		)
	if err != nil {
		t.Fatalf(
			"createTemporaryNamespaceKubeconfig() error = %v",
			err,
		)
	}

	if contextName != "staging" {
		t.Fatalf(
			"expected staging context, got %q",
			contextName,
		)
	}

	temporaryConfig, err := clientcmd.
		LoadFromFile(temporaryPath)
	if err != nil {
		t.Fatalf("load temporary kubeconfig: %v", err)
	}

	if got := temporaryConfig.
		Contexts["staging"].
		Namespace; got != "traefik" {
		t.Fatalf(
			"expected temporary namespace traefik, got %q",
			got,
		)
	}

	if got := temporaryConfig.
		Contexts["production"].
		Namespace; got != "production" {
		t.Fatalf(
			"expected production context to remain unchanged, got %q",
			got,
		)
	}

	cluster := temporaryConfig.Clusters["cluster"]

	if cluster.CertificateAuthority != "" {
		t.Fatalf(
			"expected certificate path to be flattened, got %q",
			cluster.CertificateAuthority,
		)
	}

	if string(cluster.CertificateAuthorityData) !=
		"certificate-data" {
		t.Fatalf(
			"expected embedded certificate data, got %q",
			string(cluster.CertificateAuthorityData),
		)
	}

	sourceConfig, err := clientcmd.LoadFromFile(sourcePath)
	if err != nil {
		t.Fatalf("reload source kubeconfig: %v", err)
	}

	if got := sourceConfig.
		Contexts["staging"].
		Namespace; got != "default" {
		t.Fatalf(
			"source kubeconfig was modified: namespace = %q",
			got,
		)
	}
}

func TestSetEnvironmentValueReplacesExistingValue(
	t *testing.T,
) {
	t.Parallel()

	environment := setEnvironmentValue(
		[]string{
			"PATH=/usr/bin",
			"KUBECONFIG=/old/config",
		},
		"KUBECONFIG",
		"/temporary/config",
	)

	joined := strings.Join(environment, "\n")

	if strings.Contains(joined, "/old/config") {
		t.Fatalf(
			"old KUBECONFIG remains in environment: %v",
			environment,
		)
	}

	if !strings.Contains(
		joined,
		"KUBECONFIG=/temporary/config",
	) {
		t.Fatalf(
			"temporary KUBECONFIG missing: %v",
			environment,
		)
	}
}
