package kubernetes

import (
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

var defaultContainerShells = []string{
	"/bin/bash",
	"/bin/sh",
}

func ExecPodShell(
	ctx context.Context,
	client *Client,
	namespace string,
	podName string,
	containerName string,
	requestedShell string,
) error {
	if client == nil || client.Clientset == nil || client.RESTConfig == nil {
		return fmt.Errorf("Kubernetes exec client is not configured")
	}

	shells := defaultContainerShells
	if requestedShell != "" {
		shells = []string{requestedShell}
	}

	var lastErr error
	for index, shell := range shells {
		err := execPodCommand(
			ctx,
			client,
			namespace,
			podName,
			containerName,
			[]string{shell},
		)
		if err == nil {
			return nil
		}

		lastErr = err
		if requestedShell != "" ||
			index == len(shells)-1 ||
			!isExecutableNotFoundError(err) {
			return fmt.Errorf(
				"exec %q in Pod %q container %q: %w",
				shell,
				podName,
				containerName,
				err,
			)
		}
	}

	return lastErr
}

func execPodCommand(
	ctx context.Context,
	client *Client,
	namespace string,
	podName string,
	containerName string,
	command []string,
) error {
	request := client.Clientset.
		CoreV1().
		RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(
			&corev1.PodExecOptions{
				Container: containerName,
				Command:   command,
				Stdin:     true,
				Stdout:    true,
				Stderr:    true,
				TTY:       true,
			},
			scheme.ParameterCodec,
		)

	executor, err := remotecommand.NewSPDYExecutor(
		client.RESTConfig,
		"POST",
		request.URL(),
	)
	if err != nil {
		return fmt.Errorf("create remote executor: %w", err)
	}

	stdinFD := int(os.Stdin.Fd())
	var oldState *term.State

	if term.IsTerminal(stdinFD) {
		oldState, err = term.MakeRaw(stdinFD)
		if err != nil {
			return fmt.Errorf("configure local terminal: %w", err)
		}
		defer term.Restore(stdinFD, oldState)
	}

	return executor.StreamWithContext(
		ctx,
		remotecommand.StreamOptions{
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Tty:    true,
		},
	)
}

func isExecutableNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	indicators := []string{
		"executable file not found",
		"no such file or directory",
		"command not found",
		"not found in $path",
	}

	for _, indicator := range indicators {
		if strings.Contains(message, indicator) {
			return true
		}
	}

	return false
}
