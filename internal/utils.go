package app

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func (app *App) confirmAction() error {
	color.Yellow("Confirm execution of '%s' on pod '%s' in container '%s'?", app.filePath, app.pod.Name, app.containerName)
	fmt.Print("Enter 'yes' to continue: ")
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return fmt.Errorf("error reading confirmation: %w", err)
	}
	if response != "yes" {
		return fmt.Errorf("action aborted by user")
	}
	return nil
}

func (app *App) selectOrSetContainer() (string, error) {
	if app.containerName != "" {
		return app.containerName, nil
	}

	containers := app.pod.Spec.Containers
	if len(containers) == 1 {
		return containers[0].Name, nil
	}

	fmt.Println("Multiple containers found in the pod. Please select one:")
	for i, container := range containers {
		fmt.Printf("%d. %s\n", i+1, container.Name)
	}

	var selection int
	if _, err := fmt.Scanf("%d", &selection); err != nil {
		return "", fmt.Errorf("invalid container selection: %w", err)
	}

	if selection < 1 || selection > len(containers) {
		return "", fmt.Errorf("invalid selection: %d", selection)
	}
	return containers[selection-1].Name, nil
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
