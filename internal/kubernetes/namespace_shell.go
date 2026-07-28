package kubernetes

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	namespaceShellEnvironment        = "KUBECTL_PEEK_SHELL"
	namespaceShellSessionEnvironment = "KUBECTL_PEEK_SESSION_DIR"

	namespaceShellDirectoryPrefix = "kubectl-peek-namespace-"
	namespaceShellMarkerFile      = ".kubectl-peek-session"
	namespaceShellMarkerContent   = "kubectl-peek namespace shell\n"

	namespaceShellGenerationsDirectory = "generations"
	namespaceShellCurrentLink          = "current"

	namespaceShellKubeconfigFile      = "kubeconfig"
	namespaceShellContextFile         = "context"
	namespaceShellNamespaceFile       = "namespace"
	namespaceShellPromptContextFile   = "prompt-context"
	namespaceShellPromptNamespaceFile = "prompt-namespace"
)

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
	if err := EnsureNoActiveNamespaceShell(); err != nil {
		return err
	}

	sessionDirectory, err := createNamespaceShellSessionDirectory()
	if err != nil {
		return err
	}
	defer os.RemoveAll(sessionDirectory)

	temporaryKubeconfig, targetContext, err :=
		activateNamespaceShellGeneration(
			kubeconfig,
			contextName,
			namespace,
			sessionDirectory,
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
		sessionDirectory,
		targetContext,
		namespace,
		temporaryKubeconfig,
	)
	if err != nil {
		return err
	}

	renderNamespaceShellStatus(
		out,
		"kubectl-peek namespace shell",
		targetContext,
		namespace,
		"Type `exit` to return to the previous shell",
	)

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

func SwitchNamespaceShell(
	kubeconfig string,
	contextName string,
	namespace string,
	out io.Writer,
) error {
	if !IsNamespaceShellActive() {
		return fmt.Errorf(
			"no active kubectl-peek namespace shell was found",
		)
	}

	sessionDirectory, err := activeNamespaceShellSessionDirectory()
	if err != nil {
		return err
	}

	_, targetContext, err := activateNamespaceShellGeneration(
		kubeconfig,
		contextName,
		namespace,
		sessionDirectory,
	)
	if err != nil {
		return err
	}

	renderNamespaceShellStatus(
		out,
		"kubectl-peek namespace shell switched",
		targetContext,
		namespace,
		"Continue working in the current shell",
	)

	return nil
}

func IsNamespaceShellActive() bool {
	return os.Getenv(namespaceShellEnvironment) != ""
}

func createNamespaceShellSessionDirectory() (string, error) {
	sessionDirectory, err := os.MkdirTemp(
		"",
		namespaceShellDirectoryPrefix+"*",
	)
	if err != nil {
		return "", fmt.Errorf(
			"create temporary namespace directory: %w",
			err,
		)
	}

	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(sessionDirectory)
		}
	}()

	markerPath := filepath.Join(
		sessionDirectory,
		namespaceShellMarkerFile,
	)

	if err := os.WriteFile(
		markerPath,
		[]byte(namespaceShellMarkerContent),
		0o600,
	); err != nil {
		return "", fmt.Errorf(
			"write namespace shell session marker: %w",
			err,
		)
	}

	cleanup = false
	return sessionDirectory, nil
}

func activeNamespaceShellSessionDirectory() (string, error) {
	rawSessionDirectory := strings.TrimSpace(
		os.Getenv(namespaceShellSessionEnvironment),
	)
	if rawSessionDirectory == "" {
		return "", fmt.Errorf(
			"active kubectl-peek shell has no session directory",
		)
	}

	sessionDirectory, err := filepath.Abs(rawSessionDirectory)
	if err != nil {
		return "", fmt.Errorf(
			"resolve namespace shell session directory: %w",
			err,
		)
	}

	resolvedSessionDirectory, err := filepath.EvalSymlinks(
		sessionDirectory,
	)
	if err != nil {
		return "", fmt.Errorf(
			"resolve namespace shell session directory %q: %w",
			sessionDirectory,
			err,
		)
	}

	temporaryDirectory, err := filepath.Abs(os.TempDir())
	if err != nil {
		return "", fmt.Errorf(
			"resolve operating system temporary directory: %w",
			err,
		)
	}

	resolvedTemporaryDirectory, err := filepath.EvalSymlinks(
		temporaryDirectory,
	)
	if err != nil {
		return "", fmt.Errorf(
			"resolve operating system temporary directory: %w",
			err,
		)
	}

	if filepath.Dir(resolvedSessionDirectory) !=
		resolvedTemporaryDirectory ||
		!strings.HasPrefix(
			filepath.Base(resolvedSessionDirectory),
			namespaceShellDirectoryPrefix,
		) {
		return "", fmt.Errorf(
			"invalid kubectl-peek shell session directory %q",
			sessionDirectory,
		)
	}

	info, err := os.Stat(resolvedSessionDirectory)
	if err != nil {
		return "", fmt.Errorf(
			"inspect namespace shell session directory: %w",
			err,
		)
	}

	if !info.IsDir() {
		return "", fmt.Errorf(
			"namespace shell session path %q is not a directory",
			resolvedSessionDirectory,
		)
	}

	markerPath := filepath.Join(
		resolvedSessionDirectory,
		namespaceShellMarkerFile,
	)

	marker, err := os.ReadFile(markerPath)
	if err != nil {
		return "", fmt.Errorf(
			"read namespace shell session marker: %w",
			err,
		)
	}

	if string(marker) != namespaceShellMarkerContent {
		return "", fmt.Errorf(
			"invalid namespace shell session marker in %q",
			resolvedSessionDirectory,
		)
	}

	return resolvedSessionDirectory, nil
}

func activateNamespaceShellGeneration(
	kubeconfig string,
	contextName string,
	namespace string,
	sessionDirectory string,
) (string, string, error) {
	temporaryConfig, targetContext, err :=
		temporaryNamespaceConfig(
			kubeconfig,
			contextName,
			namespace,
		)
	if err != nil {
		return "", "", err
	}

	generationsDirectory := filepath.Join(
		sessionDirectory,
		namespaceShellGenerationsDirectory,
	)

	if err := os.MkdirAll(
		generationsDirectory,
		0o700,
	); err != nil {
		return "", "", fmt.Errorf(
			"create namespace shell generations directory: %w",
			err,
		)
	}

	generationDirectory, err := os.MkdirTemp(
		generationsDirectory,
		"generation-*",
	)
	if err != nil {
		return "", "", fmt.Errorf(
			"create namespace shell generation: %w",
			err,
		)
	}

	activated := false
	defer func() {
		if !activated {
			_ = os.RemoveAll(generationDirectory)
		}
	}()

	kubeconfigPath := filepath.Join(
		generationDirectory,
		namespaceShellKubeconfigFile,
	)

	if err := writeNamespaceShellKubeconfig(
		kubeconfigPath,
		temporaryConfig,
		targetContext,
		namespace,
	); err != nil {
		return "", "", err
	}

	stateFiles := map[string]string{
		namespaceShellContextFile:         targetContext,
		namespaceShellNamespaceFile:       namespace,
		namespaceShellPromptContextFile:   sanitizePromptValue(targetContext),
		namespaceShellPromptNamespaceFile: sanitizePromptValue(namespace),
	}

	for name, value := range stateFiles {
		if err := writeNamespaceShellStateValue(
			filepath.Join(generationDirectory, name),
			value,
		); err != nil {
			return "", "", err
		}
	}

	relativeGeneration, err := filepath.Rel(
		sessionDirectory,
		generationDirectory,
	)
	if err != nil {
		return "", "", fmt.Errorf(
			"resolve namespace shell generation link target: %w",
			err,
		)
	}

	nextLink := filepath.Join(
		sessionDirectory,
		".current-"+filepath.Base(generationDirectory),
	)
	defer os.Remove(nextLink)

	if err := os.Symlink(
		relativeGeneration,
		nextLink,
	); err != nil {
		return "", "", fmt.Errorf(
			"create namespace shell generation link: %w",
			err,
		)
	}

	currentLink := filepath.Join(
		sessionDirectory,
		namespaceShellCurrentLink,
	)

	if err := os.Rename(
		nextLink,
		currentLink,
	); err != nil {
		return "", "", fmt.Errorf(
			"activate namespace shell generation: %w",
			err,
		)
	}

	activated = true

	return filepath.Join(
		currentLink,
		namespaceShellKubeconfigFile,
	), targetContext, nil
}

func temporaryNamespaceConfig(
	kubeconfig string,
	contextName string,
	namespace string,
) (*clientcmdapi.Config, string, error) {
	loadingRules := newLoadingRules(kubeconfig)

	config, err := loadingRules.Load()
	if err != nil {
		return nil, "", fmt.Errorf(
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
		return nil, "", err
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
		return nil, "", fmt.Errorf(
			"flatten temporary Kubernetes configuration: %w",
			err,
		)
	}

	return temporaryConfig, targetContext, nil
}

func writeNamespaceShellKubeconfig(
	path string,
	config *clientcmdapi.Config,
	targetContext string,
	namespace string,
) error {
	data, err := clientcmd.Write(*config)
	if err != nil {
		return fmt.Errorf(
			"encode temporary Kubernetes configuration: %w",
			err,
		)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf(
			"write temporary Kubernetes configuration: %w",
			err,
		)
	}

	loaded, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return fmt.Errorf(
			"validate temporary Kubernetes configuration: %w",
			err,
		)
	}

	if loaded.CurrentContext != targetContext {
		return fmt.Errorf(
			"temporary Kubernetes configuration uses context %q, expected %q",
			loaded.CurrentContext,
			targetContext,
		)
	}

	context, found := loaded.Contexts[targetContext]
	if !found || context == nil {
		return fmt.Errorf(
			"temporary Kubernetes configuration does not contain context %q",
			targetContext,
		)
	}

	if context.Namespace != namespace {
		return fmt.Errorf(
			"temporary Kubernetes configuration uses namespace %q, expected %q",
			context.Namespace,
			namespace,
		)
	}

	return nil
}

func writeNamespaceShellStateValue(
	path string,
	value string,
) error {
	if value == "" {
		return fmt.Errorf(
			"namespace shell state value for %q is empty",
			filepath.Base(path),
		)
	}

	if strings.ContainsAny(value, "\r\n\x00") {
		return fmt.Errorf(
			"namespace shell state value for %q contains unsupported characters",
			filepath.Base(path),
		)
	}

	if err := os.WriteFile(
		path,
		[]byte(value+"\n"),
		0o600,
	); err != nil {
		return fmt.Errorf(
			"write namespace shell state %q: %w",
			filepath.Base(path),
			err,
		)
	}

	return nil
}

func sanitizePromptValue(value string) string {
	var builder strings.Builder

	for _, character := range value {
		if unicode.IsLetter(character) ||
			unicode.IsNumber(character) ||
			strings.ContainsRune("._-:/@", character) {
			builder.WriteRune(character)
			continue
		}

		builder.WriteRune('?')
	}

	return builder.String()
}

func createTemporaryNamespaceKubeconfig(
	kubeconfig string,
	contextName string,
	namespace string,
	temporaryDirectory string,
) (string, string, error) {
	temporaryConfig, targetContext, err :=
		temporaryNamespaceConfig(
			kubeconfig,
			contextName,
			namespace,
		)
	if err != nil {
		return "", "", err
	}

	path := filepath.Join(
		temporaryDirectory,
		namespaceShellKubeconfigFile,
	)

	if err := writeNamespaceShellKubeconfig(
		path,
		temporaryConfig,
		targetContext,
		namespace,
	); err != nil {
		return "", "", err
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
	sessionDirectory string,
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
		namespaceShellSessionEnvironment,
		sessionDirectory,
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
			sessionDirectory,
			environment,
		)
		return command, updatedEnvironment, err

	case "bash":
		command, err := bashNamespaceShell(
			shellPath,
			sessionDirectory,
		)
		return command, environment, err

	default:
		return exec.Command(shellPath, "-i"), environment, nil
	}
}

func zshNamespaceShell(
	shellPath string,
	sessionDirectory string,
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
		sessionDirectory,
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

	configuration.WriteString(`
typeset -g KUBECTL_PEEK_BASE_PROMPT="$PROMPT"

_kubectl_peek_refresh_prompt() {
  local context namespace prompt_context prompt_namespace

  if ! IFS= read -r context < "$KUBECTL_PEEK_SESSION_DIR/current/context"; then
    context="?"
  fi

  if ! IFS= read -r namespace < "$KUBECTL_PEEK_SESSION_DIR/current/namespace"; then
    namespace="?"
  fi

  if ! IFS= read -r prompt_context < "$KUBECTL_PEEK_SESSION_DIR/current/prompt-context"; then
    prompt_context="?"
  fi

  if ! IFS= read -r prompt_namespace < "$KUBECTL_PEEK_SESSION_DIR/current/prompt-namespace"; then
    prompt_namespace="?"
  fi

  export KUBECTL_PEEK_CONTEXT="$context"
  export KUBECTL_PEEK_NAMESPACE="$namespace"

  PROMPT="%B%F{cyan}[k8s:${prompt_context} %F{yellow}ns:${prompt_namespace}%F{cyan}]%f%b ${KUBECTL_PEEK_BASE_PROMPT}"
}

autoload -Uz add-zsh-hook
add-zsh-hook precmd _kubectl_peek_refresh_prompt
_kubectl_peek_refresh_prompt
`)

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
	sessionDirectory string,
) (*exec.Cmd, error) {
	bashrc := filepath.Join(
		sessionDirectory,
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

	configuration.WriteString(`
KUBECTL_PEEK_BASE_PS1="$PS1"

_kubectl_peek_refresh_prompt() {
  local context namespace prompt_context prompt_namespace

  if ! IFS= read -r context < "$KUBECTL_PEEK_SESSION_DIR/current/context"; then
    context="?"
  fi

  if ! IFS= read -r namespace < "$KUBECTL_PEEK_SESSION_DIR/current/namespace"; then
    namespace="?"
  fi

  if ! IFS= read -r prompt_context < "$KUBECTL_PEEK_SESSION_DIR/current/prompt-context"; then
    prompt_context="?"
  fi

  if ! IFS= read -r prompt_namespace < "$KUBECTL_PEEK_SESSION_DIR/current/prompt-namespace"; then
    prompt_namespace="?"
  fi

  export KUBECTL_PEEK_CONTEXT="$context"
  export KUBECTL_PEEK_NAMESPACE="$namespace"

  PS1="\[\033[1;36m\][k8s:${prompt_context} \[\033[1;33m\]ns:${prompt_namespace}\[\033[1;36m\]]\[\033[0m\] ${KUBECTL_PEEK_BASE_PS1}"
}

if declare -p PROMPT_COMMAND 2>/dev/null | command grep -q '^declare -a'; then
  PROMPT_COMMAND+=(_kubectl_peek_refresh_prompt)
else
  KUBECTL_PEEK_ORIGINAL_PROMPT_COMMAND="${PROMPT_COMMAND-}"

  if [[ -n "$KUBECTL_PEEK_ORIGINAL_PROMPT_COMMAND" ]]; then
    PROMPT_COMMAND="${KUBECTL_PEEK_ORIGINAL_PROMPT_COMMAND};_kubectl_peek_refresh_prompt"
  else
    PROMPT_COMMAND="_kubectl_peek_refresh_prompt"
  fi
fi

_kubectl_peek_refresh_prompt
`)

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

func renderNamespaceShellStatus(
	out io.Writer,
	title string,
	contextName string,
	namespace string,
	footer string,
) {
	fmt.Fprintln(out)
	fmt.Fprintf(
		out,
		"%s%s┌─ %s%s\n",
		ansiBold,
		ansiCyan,
		title,
		ansiReset,
	)
	fmt.Fprintf(
		out,
		"%s│%s Context    %s%s%s\n",
		ansiCyan,
		ansiReset,
		ansiBold,
		contextName,
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
		"%s└─%s %s%s%s\n",
		ansiCyan,
		ansiReset,
		ansiDim,
		footer,
		ansiReset,
	)
	fmt.Fprintln(out)
}

func readNamespaceShellState(
	sessionDirectory string,
) (string, string, error) {
	currentDirectory := filepath.Join(
		sessionDirectory,
		namespaceShellCurrentLink,
	)

	contextName, err := readNamespaceShellStateValue(
		filepath.Join(
			currentDirectory,
			namespaceShellContextFile,
		),
	)
	if err != nil {
		return "", "", err
	}

	namespace, err := readNamespaceShellStateValue(
		filepath.Join(
			currentDirectory,
			namespaceShellNamespaceFile,
		),
	)
	if err != nil {
		return "", "", err
	}

	return contextName, namespace, nil
}

func readNamespaceShellStateValue(
	path string,
) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf(
			"read namespace shell state %q: %w",
			filepath.Base(path),
			err,
		)
	}

	value := strings.TrimSuffix(
		strings.TrimSuffix(string(data), "\n"),
		"\r",
	)
	if value == "" {
		return "", fmt.Errorf(
			"namespace shell state %q is empty",
			filepath.Base(path),
		)
	}

	return value, nil
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

func EnsureNoActiveNamespaceShell() error {
	if os.Getenv(namespaceShellEnvironment) == "" {
		return nil
	}

	currentContext := os.Getenv("KUBECTL_PEEK_CONTEXT")
	currentNamespace := os.Getenv("KUBECTL_PEEK_NAMESPACE")

	if currentContext != "" && currentNamespace != "" {
		return fmt.Errorf(
			"an isolated kubectl-peek shell is already active "+
				"(context %q, namespace %q); run exit before opening another one",
			currentContext,
			currentNamespace,
		)
	}

	return fmt.Errorf(
		"an isolated kubectl-peek shell is already active; " +
			"run exit before opening another one",
	)
}
