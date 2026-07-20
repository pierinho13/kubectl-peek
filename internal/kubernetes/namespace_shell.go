package kubernetes

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const namespaceShellEnvironment = "KUBECTL_PEEK_SHELL"

const (
	ansiBold   = "\033[1m"
	ansiCyan   = "\033[36m"
	ansiYellow = "\033[33m"
	ansiDim    = "\033[2m"
	ansiReset  = "\033[0m"
)

func RunNamespaceShell(
	kubeconfig string,
	contextName string,
	namespace string,
	out io.Writer,
) error {
	if os.Getenv(namespaceShellEnvironment) != "" {
		return fmt.Errorf(
			"a kubectl-peek namespace shell is already active; run exit before opening another one",
		)
	}

	temporaryDirectory, err := os.MkdirTemp(
		"",
		"kubectl-peek-namespace-*",
	)
	if err != nil {
		return fmt.Errorf(
			"create temporary namespace directory: %w",
			err,
		)
	}
	defer os.RemoveAll(temporaryDirectory)

	temporaryKubeconfig, targetContext, err :=
		createTemporaryNamespaceKubeconfig(
			kubeconfig,
			contextName,
			namespace,
			temporaryDirectory,
		)
	if err != nil {
		return err
	}

	shellPath, err := resolveInteractiveShell()
	if err != nil {
		return err
	}

	command, environment, err := namespaceShellCommand(
		shellPath,
		temporaryDirectory,
		targetContext,
		namespace,
		temporaryKubeconfig,
	)
	if err != nil {
		return err
	}

	fmt.Fprintln(out)
	fmt.Fprintf(
		out,
		"%s%s┌─ kubectl-peek namespace shell%s\n",
		ansiBold,
		ansiCyan,
		ansiReset,
	)
	fmt.Fprintf(
		out,
		"%s│%s Context    %s%s%s\n",
		ansiCyan,
		ansiReset,
		ansiBold,
		targetContext,
		ansiReset,
	)
	fmt.Fprintf(
		out,
		"%s│%s Namespace  %s%s%s%s\n",
		ansiCyan,
		ansiReset,
		ansiBold,
		ansiYellow,
		namespace,
		ansiReset,
	)
	fmt.Fprintf(
		out,
		"%s└─%s %sType `exit` to return to the previous shell%s\n",
		ansiCyan,
		ansiReset,
		ansiDim,
		ansiReset,
	)
	fmt.Fprintln(out)

	command.Env = environment
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		return fmt.Errorf(
			"run namespace shell %q: %w",
			shellPath,
			err,
		)
	}

	return nil
}

func createTemporaryNamespaceKubeconfig(
	kubeconfig string,
	contextName string,
	namespace string,
	temporaryDirectory string,
) (string, string, error) {
	loadingRules := newLoadingRules(kubeconfig)

	config, err := loadingRules.Load()
	if err != nil {
		return "", "", fmt.Errorf(
			"load Kubernetes configuration: %w",
			err,
		)
	}

	targetContext, err := resolveContextName(
		config.CurrentContext,
		config.Contexts,
		contextName,
	)
	if err != nil {
		return "", "", err
	}

	// Work on a copy. The user's source kubeconfig is never modified.
	temporaryConfig := config.DeepCopy()
	temporaryConfig.CurrentContext = targetContext

	updatedContext := temporaryConfig.
		Contexts[targetContext].
		DeepCopy()
	updatedContext.Namespace = namespace
	temporaryConfig.Contexts[targetContext] = updatedContext

	// Flatten path-based certificate and key references into the temporary
	// kubeconfig. This preserves complex merged kubeconfigs even when their
	// original files contain relative paths.
	if err := clientcmdapi.FlattenConfig(temporaryConfig); err != nil {
		return "", "", fmt.Errorf(
			"flatten temporary Kubernetes configuration: %w",
			err,
		)
	}

	data, err := clientcmd.Write(*temporaryConfig)
	if err != nil {
		return "", "", fmt.Errorf(
			"encode temporary Kubernetes configuration: %w",
			err,
		)
	}

	path := filepath.Join(
		temporaryDirectory,
		"kubeconfig",
	)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", "", fmt.Errorf(
			"write temporary Kubernetes configuration: %w",
			err,
		)
	}

	return path, targetContext, nil
}

func resolveContextName(
	currentContext string,
	contexts map[string]*clientcmdapi.Context,
	requestedContext string,
) (string, error) {
	targetContext := requestedContext

	if targetContext == "" {
		targetContext = currentContext
	}

	if targetContext == "" {
		return "", fmt.Errorf(
			"Kubernetes configuration has no current context",
		)
	}

	context, ok := contexts[targetContext]
	if !ok || context == nil {
		return "", fmt.Errorf(
			"Kubernetes context %q not found",
			targetContext,
		)
	}

	return targetContext, nil
}

func resolveInteractiveShell() (string, error) {
	shellPath := strings.TrimSpace(os.Getenv("SHELL"))

	if shellPath == "" {
		if runtime.GOOS == "windows" {
			return "", fmt.Errorf(
				"namespace shell is not supported on Windows yet",
			)
		}

		shellPath = "/bin/sh"
	}

	if _, err := os.Stat(shellPath); err != nil {
		return "", fmt.Errorf(
			"interactive shell %q is unavailable: %w",
			shellPath,
			err,
		)
	}

	return shellPath, nil
}

func namespaceShellCommand(
	shellPath string,
	temporaryDirectory string,
	contextName string,
	namespace string,
	temporaryKubeconfig string,
) (*exec.Cmd, []string, error) {
	environment := append(
		[]string(nil),
		os.Environ()...,
	)

	environment = setEnvironmentValue(
		environment,
		"KUBECONFIG",
		temporaryKubeconfig,
	)
	environment = setEnvironmentValue(
		environment,
		namespaceShellEnvironment,
		"1",
	)
	environment = setEnvironmentValue(
		environment,
		"KUBECTL_PEEK_CONTEXT",
		contextName,
	)
	environment = setEnvironmentValue(
		environment,
		"KUBECTL_PEEK_NAMESPACE",
		namespace,
	)

	switch filepath.Base(shellPath) {
	case "zsh":
		command, updatedEnvironment, err := zshNamespaceShell(
			shellPath,
			temporaryDirectory,
			contextName,
			namespace,
			environment,
		)
		return command, updatedEnvironment, err

	case "bash":
		command, err := bashNamespaceShell(
			shellPath,
			temporaryDirectory,
			contextName,
			namespace,
		)
		return command, environment, err

	default:
		return exec.Command(shellPath, "-i"), environment, nil
	}
}

func zshNamespaceShell(
	shellPath string,
	temporaryDirectory string,
	contextName string,
	namespace string,
	environment []string,
) (*exec.Cmd, []string, error) {
	originalZDOTDIR := os.Getenv("ZDOTDIR")

	if originalZDOTDIR == "" {
		homeDirectory, err := os.UserHomeDir()
		if err != nil {
			return nil, nil, fmt.Errorf(
				"resolve home directory: %w",
				err,
			)
		}

		originalZDOTDIR = homeDirectory
	}

	zshDirectory := filepath.Join(
		temporaryDirectory,
		"zsh",
	)

	if err := os.MkdirAll(zshDirectory, 0o700); err != nil {
		return nil, nil, fmt.Errorf(
			"create temporary zsh configuration: %w",
			err,
		)
	}

	originalZshrc := filepath.Join(
		originalZDOTDIR,
		".zshrc",
	)

	var configuration strings.Builder

	if _, err := os.Stat(originalZshrc); err == nil {
		fmt.Fprintf(
			&configuration,
			"source %s\n",
			shellQuote(originalZshrc),
		)
	}

	fmt.Fprintf(
		&configuration,
		"PROMPT='%%B%%F{cyan}[k8s:%s %%F{yellow}ns:%s%%F{cyan}]%%f%%b '\"$PROMPT\"\n",
		contextName,
		namespace,
	)

	zshrc := filepath.Join(zshDirectory, ".zshrc")

	if err := os.WriteFile(
		zshrc,
		[]byte(configuration.String()),
		0o600,
	); err != nil {
		return nil, nil, fmt.Errorf(
			"write temporary zsh configuration: %w",
			err,
		)
	}

	environment = setEnvironmentValue(
		environment,
		"ZDOTDIR",
		zshDirectory,
	)

	return exec.Command(shellPath, "-i"), environment, nil
}

func bashNamespaceShell(
	shellPath string,
	temporaryDirectory string,
	contextName string,
	namespace string,
) (*exec.Cmd, error) {
	bashrc := filepath.Join(
		temporaryDirectory,
		"bashrc",
	)

	var configuration strings.Builder

	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf(
			"resolve home directory: %w",
			err,
		)
	}

	originalBashrc := filepath.Join(
		homeDirectory,
		".bashrc",
	)

	if _, err := os.Stat(originalBashrc); err == nil {
		fmt.Fprintf(
			&configuration,
			"source %s\n",
			shellQuote(originalBashrc),
		)
	}

	fmt.Fprintf(
		&configuration,
		"PS1='\\[\\033[1;36m\\][k8s:%s \\[\\033[1;33m\\]ns:%s\\[\\033[1;36m\\]]\\[\\033[0m\\] '\"$PS1\"\n",
		contextName,
		namespace,
	)

	if err := os.WriteFile(
		bashrc,
		[]byte(configuration.String()),
		0o600,
	); err != nil {
		return nil, fmt.Errorf(
			"write temporary bash configuration: %w",
			err,
		)
	}

	return exec.Command(
		shellPath,
		"--rcfile",
		bashrc,
		"-i",
	), nil
}

func setEnvironmentValue(
	environment []string,
	key string,
	value string,
) []string {
	prefix := key + "="
	result := make([]string, 0, len(environment)+1)

	for _, entry := range environment {
		if strings.HasPrefix(entry, prefix) {
			continue
		}

		result = append(result, entry)
	}

	return append(result, prefix+value)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(
		value,
		"'",
		`'\''`,
	) + "'"
}
