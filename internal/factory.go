package app

import (
	"flag"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	kubeContext   string
	filePath      string
	podName       string
	containerName string
	confirm       bool
	fileType      string
	args          string
)

type App struct {
	kubeContext   string
	filePath      string
	podName       string
	containerName string
	noConfirm     bool
	fileType      string
	args          string
	clientset     *kubernetes.Clientset
	config        *rest.Config
	pod           *corev1.Pod
}

// Option setters for App struct
func WithKubeContext(context string) func(app *App) {
	return func(app *App) {
		app.kubeContext = context
	}
}

func WithFilePath(filePath string) func(app *App) {
	return func(app *App) {
		app.filePath = filePath
	}
}

func WithPodName(podName string) func(app *App) {
	return func(app *App) {
		app.podName = podName
	}
}

func WithContainerName(containerName string) func(app *App) {
	return func(app *App) {
		app.containerName = containerName
	}
}

func WithNoConfirm(noConfirm bool) func(app *App) {
	return func(app *App) {
		app.noConfirm = noConfirm
	}
}

func WithFileType(fileType string) func(app *App) {
	return func(app *App) {
		app.fileType = fileType
	}
}

func WithArgs(args string) func(app *App) {
	return func(app *App) {
		app.args = args
	}
}

// Create a new App instance and validate required fields
func NewApp(opts ...func(app *App)) *App {
	app := &App{}
	for _, opt := range opts {
		opt(app)
	}

	app.validateRequiredFields()

	return app
}

func (app *App) validateRequiredFields() {
	if app.kubeContext == "" || app.filePath == "" || app.podName == "" {
		fmt.Println("Error: context, file, and pod name are required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if app.fileType != "auto" && app.fileType != "script" && app.fileType != "binary" {
		fmt.Println("Error: type must be 'auto', 'script', or 'binary'")
		os.Exit(1)
	}
}
