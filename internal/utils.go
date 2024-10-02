package app

import (
	"context"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func (app *App) selectOrSetContainer() error {
	if app.containerName != "" {
		return nil
	}

	containers := app.pod.Spec.Containers
	if len(containers) == 1 {
		app.containerName = containers[0].Name
	} else {
		containersNames := make([]string, 0)
		for _, container := range containers {
			containersNames = append(containersNames, container.Name)
		}
		containerName, err := runContainerSelection(containersNames)
		if err != nil {
			return fmt.Errorf("error running container selection: %w", err)
		}
		app.containerName = containerName
	}

	return nil
}

func (app *App) runCommandInPod(ctx context.Context, command []string) error {
	req := app.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(app.pod.Name).
		Namespace(app.pod.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: app.containerName,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(app.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("error creating SPDY executor: %w", err)
	}

	return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
