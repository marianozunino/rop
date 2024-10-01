package app

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
)

// Run executes the main logic of the application
func (app *App) Run(ctx context.Context) error {
	color.Green("Executing file: %s", app.filePath)

	if err := app.validateFileExistence(); err != nil {
		return err
	}

	if err := app.initializeKubernetesClient(); err != nil {
		return err
	}

	if app.noConfirm {
		color.Magenta("Running without confirmation in 5 seconds...Press CTRL+C to cancel")
		time.Sleep(5 * time.Second)
	}

	if err := app.setupPodAndContainer(ctx); err != nil {
		return err
	}

	if !app.noConfirm {
		if err := app.confirmAction(fmt.Sprintf("Confirm execution of '%s' on pod '%s' in container '%s'?", app.filePath, app.pod.Name, app.containerName)); err != nil {
			return err
		}
	}

	color.Yellow("Executing %s inside %s (container: %s)...", app.filePath, app.pod.Name, app.containerName)

	return app.executeFile(ctx)
}

func (app *App) initializeKubernetesClient() error {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	fmt.Printf("Using kubeconfig: %s\n", kubeconfig)

	config, err := buildConfigWithContextFromFlags(app.kubeContext, kubeconfig)
	if err != nil {
		return fmt.Errorf("error creating Kubernetes client: %w", err)
	}
	app.config = config

	clientset, err := kubernetes.NewForConfig(app.config)
	if err != nil {
		return fmt.Errorf("error creating Kubernetes clientset: %w", err)
	}
	app.clientset = clientset

	return nil
}

func buildConfigWithContextFromFlags(context, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

func (app *App) validateFileExistence() error {
	if _, err := os.Stat(app.filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", app.filePath)
	}
	return nil
}

func (app *App) confirmAction(message string) error {
	color.Yellow(message)
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

func (app *App) setupPodAndContainer(ctx context.Context) error {
	pod, err := app.getRunningPod(ctx)
	if err != nil {
		return err
	}
	app.pod = pod

	containerName, err := app.selectOrSetContainer()
	if err != nil {
		return err
	}
	app.containerName = containerName

	return nil
}

func (app *App) getRunningPod(ctx context.Context) (*corev1.Pod, error) {
	pods, err := app.clientset.CoreV1().Pods("oc-app").List(ctx, metav1.ListOptions{
		FieldSelector: "status.phase=Running",
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", app.podName),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting pod: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no running pods found for %s", app.podName)
	}
	color.Green("Found pod %s", pods.Items[0].Name)
	return &pods.Items[0], nil
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

func (app *App) executeFile(ctx context.Context) error {
	file, err := os.Open(app.filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error getting file info: %w", err)
	}

	isExecutable := fileInfo.Mode()&0o111 != 0
	if app.fileType == "auto" {
		app.fileType = "script"
		if isExecutable {
			app.fileType = "binary"
		}
	}

	tempPath := fmt.Sprintf("/tmp/%s", filepath.Base(app.filePath))

	// Ensure cleanup happens
	defer func() {
		if err := app.deleteFileFromPod(ctx, tempPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to delete file %s from pod: %v\n", tempPath, err)
		} else {
			fmt.Fprintf(os.Stdout, "Deleted file %s from pod\n", tempPath)
		}
	}()

	if err := app.copyFileToPod(ctx, file, tempPath); err != nil {
		return fmt.Errorf("failed to copy file to pod: %w", err)
	}

	if app.fileType == "script" {
		return app.executeScript(ctx, tempPath)
	}
	return app.executeBinary(ctx, tempPath)
}

func (app *App) deleteFileFromPod(ctx context.Context, filePath string) error {
	cmd := []string{"rm", "-f", filePath}
	return app.runCommandInPod(ctx, cmd)
}

func (app *App) executeScript(ctx context.Context, filePath string) error {
	command := []string{"sh", filePath}
	if filepath.Ext(app.filePath) == ".js" {
		command = []string{"node", filePath}
	}
	if app.args != "" {
		command = append(command, app.args)
	}
	return app.runCommandInPod(ctx, command)
}

func (app *App) executeBinary(ctx context.Context, filePath string) error {
	command := []string{filePath}
	if app.args != "" {
		command = append(command, app.args)
	}
	return app.runCommandInPod(ctx, command)
}

func (app *App) copyFileToPod(ctx context.Context, file *os.File, destPath string) error {
	command := []string{"cp", "/dev/stdin", destPath}

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

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  file,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return fmt.Errorf("error copying file to pod: %w, stderr: %s", err, stderr.String())
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
