package kubernetes

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestActivateNamespaceShellGenerationSwitchesCurrent(
	t *testing.T,
) {
	sourcePath := writeNamespaceShellTestConfig(t)

	sessionDirectory, err :=
		createNamespaceShellSessionDirectory()
	if err != nil {
		t.Fatalf(
			"createNamespaceShellSessionDirectory() error = %v",
			err,
		)
	}
	defer os.RemoveAll(sessionDirectory)

	stableKubeconfig, targetContext, err :=
		activateNamespaceShellGeneration(
			sourcePath,
			"staging",
			"traefik",
			sessionDirectory,
		)
	if err != nil {
		t.Fatalf(
			"activate first generation: %v",
			err,
		)
	}

	if targetContext != "staging" {
		t.Fatalf(
			"first target context = %q, want staging",
			targetContext,
		)
	}

	expectedStablePath := filepath.Join(
		sessionDirectory,
		namespaceShellCurrentLink,
		namespaceShellKubeconfigFile,
	)
	if stableKubeconfig != expectedStablePath {
		t.Fatalf(
			"stable kubeconfig = %q, want %q",
			stableKubeconfig,
			expectedStablePath,
		)
	}

	firstGeneration, err := filepath.EvalSymlinks(
		filepath.Join(
			sessionDirectory,
			namespaceShellCurrentLink,
		),
	)
	if err != nil {
		t.Fatalf("resolve first generation: %v", err)
	}

	secondStableKubeconfig, targetContext, err :=
		activateNamespaceShellGeneration(
			sourcePath,
			"production",
			"monitoring",
			sessionDirectory,
		)
	if err != nil {
		t.Fatalf(
			"activate second generation: %v",
			err,
		)
	}

	if secondStableKubeconfig != stableKubeconfig {
		t.Fatalf(
			"kubeconfig path changed from %q to %q",
			stableKubeconfig,
			secondStableKubeconfig,
		)
	}

	if targetContext != "production" {
		t.Fatalf(
			"second target context = %q, want production",
			targetContext,
		)
	}

	secondGeneration, err := filepath.EvalSymlinks(
		filepath.Join(
			sessionDirectory,
			namespaceShellCurrentLink,
		),
	)
	if err != nil {
		t.Fatalf("resolve second generation: %v", err)
	}

	if firstGeneration == secondGeneration {
		t.Fatalf(
			"current link still points to first generation %q",
			firstGeneration,
		)
	}

	if _, err := os.Stat(firstGeneration); err != nil {
		t.Fatalf(
			"first generation should remain until shell exit: %v",
			err,
		)
	}

	activeConfig, err := clientcmd.LoadFromFile(
		stableKubeconfig,
	)
	if err != nil {
		t.Fatalf("load active generation: %v", err)
	}

	if activeConfig.CurrentContext != "production" {
		t.Fatalf(
			"active context = %q, want production",
			activeConfig.CurrentContext,
		)
	}

	if got := activeConfig.
		Contexts["production"].
		Namespace; got != "monitoring" {
		t.Fatalf(
			"active namespace = %q, want monitoring",
			got,
		)
	}

	contextName, namespace, err := readNamespaceShellState(
		sessionDirectory,
	)
	if err != nil {
		t.Fatalf("read active shell state: %v", err)
	}

	if contextName != "production" ||
		namespace != "monitoring" {
		t.Fatalf(
			"active state = %q/%q, want production/monitoring",
			contextName,
			namespace,
		)
	}
}

func TestSwitchNamespaceShellUpdatesActiveSession(
	t *testing.T,
) {
	sourcePath := writeNamespaceShellTestConfig(t)

	sessionDirectory, err :=
		createNamespaceShellSessionDirectory()
	if err != nil {
		t.Fatalf(
			"createNamespaceShellSessionDirectory() error = %v",
			err,
		)
	}
	defer os.RemoveAll(sessionDirectory)

	_, _, err = activateNamespaceShellGeneration(
		sourcePath,
		"staging",
		"default",
		sessionDirectory,
	)
	if err != nil {
		t.Fatalf("activate initial generation: %v", err)
	}

	t.Setenv(namespaceShellEnvironment, "1")
	t.Setenv(
		namespaceShellSessionEnvironment,
		sessionDirectory,
	)

	var output bytes.Buffer

	if err := SwitchNamespaceShell(
		sourcePath,
		"production",
		"monitoring",
		&output,
	); err != nil {
		t.Fatalf("SwitchNamespaceShell() error = %v", err)
	}

	contextName, namespace, err := readNamespaceShellState(
		sessionDirectory,
	)
	if err != nil {
		t.Fatalf("read switched shell state: %v", err)
	}

	if contextName != "production" ||
		namespace != "monitoring" {
		t.Fatalf(
			"switched state = %q/%q, want production/monitoring",
			contextName,
			namespace,
		)
	}

	rendered := output.String()
	if !strings.Contains(
		rendered,
		"namespace shell switched",
	) {
		t.Fatalf(
			"switch output does not describe the change:\n%s",
			rendered,
		)
	}
}

func TestNamespaceShellCommandUsesStableSessionPaths(
	t *testing.T,
) {
	sessionDirectory := t.TempDir()
	stableKubeconfig := filepath.Join(
		sessionDirectory,
		namespaceShellCurrentLink,
		namespaceShellKubeconfigFile,
	)

	_, environment, err := namespaceShellCommand(
		"/bin/bash",
		sessionDirectory,
		"staging",
		"traefik",
		stableKubeconfig,
	)
	if err != nil {
		t.Fatalf("namespaceShellCommand() error = %v", err)
	}

	joinedEnvironment := strings.Join(environment, "\n")

	expectedEnvironment := []string{
		"KUBECONFIG=" + stableKubeconfig,
		namespaceShellEnvironment + "=1",
		namespaceShellSessionEnvironment +
			"=" + sessionDirectory,
	}

	for _, expected := range expectedEnvironment {
		if !strings.Contains(joinedEnvironment, expected) {
			t.Fatalf(
				"environment does not contain %q:\n%s",
				expected,
				joinedEnvironment,
			)
		}
	}

	bashrc, err := os.ReadFile(
		filepath.Join(sessionDirectory, "bashrc"),
	)
	if err != nil {
		t.Fatalf("read generated bashrc: %v", err)
	}

	configuration := string(bashrc)

	for _, expected := range []string{
		"current/prompt-context",
		"current/prompt-namespace",
		"_kubectl_peek_refresh_prompt",
		"PROMPT_COMMAND",
	} {
		if !strings.Contains(configuration, expected) {
			t.Fatalf(
				"generated bashrc does not contain %q:\n%s",
				expected,
				configuration,
			)
		}
	}
}

func TestZshNamespaceShellUsesDynamicGenerationState(
	t *testing.T,
) {
	sessionDirectory := t.TempDir()
	stableKubeconfig := filepath.Join(
		sessionDirectory,
		namespaceShellCurrentLink,
		namespaceShellKubeconfigFile,
	)

	_, environment, err := namespaceShellCommand(
		"/bin/zsh",
		sessionDirectory,
		"staging",
		"traefik",
		stableKubeconfig,
	)
	if err != nil {
		t.Fatalf("namespaceShellCommand() error = %v", err)
	}

	if !strings.Contains(
		strings.Join(environment, "\n"),
		"ZDOTDIR="+filepath.Join(sessionDirectory, "zsh"),
	) {
		t.Fatalf(
			"generated environment does not contain temporary ZDOTDIR: %v",
			environment,
		)
	}

	zshrc, err := os.ReadFile(
		filepath.Join(
			sessionDirectory,
			"zsh",
			".zshrc",
		),
	)
	if err != nil {
		t.Fatalf("read generated zshrc: %v", err)
	}

	configuration := string(zshrc)

	for _, expected := range []string{
		"current/context",
		"current/namespace",
		"current/prompt-context",
		"current/prompt-namespace",
		"add-zsh-hook precmd",
		"_kubectl_peek_refresh_prompt",
	} {
		if !strings.Contains(configuration, expected) {
			t.Fatalf(
				"generated zshrc does not contain %q:\n%s",
				expected,
				configuration,
			)
		}
	}
}

func TestSanitizePromptValue(
	t *testing.T,
) {
	t.Parallel()

	got := sanitizePromptValue(
		`gke_project/cluster $(touch /tmp/bad)%`,
	)

	want := "gke_project/cluster???touch?/tmp/bad??"

	if got != want {
		t.Fatalf(
			"sanitizePromptValue() = %q, want %q",
			got,
			want,
		)
	}
}

func writeNamespaceShellTestConfig(
	t *testing.T,
) string {
	t.Helper()

	config := clientcmdapi.NewConfig()
	config.CurrentContext = "staging"
	config.Clusters["cluster"] = &clientcmdapi.Cluster{
		Server: "https://example.invalid",
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

	path := filepath.Join(
		t.TempDir(),
		"config",
	)

	if err := clientcmd.WriteToFile(*config, path); err != nil {
		t.Fatalf("write test kubeconfig: %v", err)
	}

	return path
}
