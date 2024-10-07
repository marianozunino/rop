package app

import (
	"flag"
	"os"

	"github.com/marianozunino/rop/internal/k8s"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
)

type App struct {
	filePath  string
	podName   string
	noConfirm bool
	fileType  string
	args      []string
	destPath  string
	runner    string

	client      *k8s.Client
	kubeContext string
	namespace   string
	pod         *corev1.Pod
	container   string
}

// Option setters for App struct
func WithKubeContext(context string) func(app *App) {
	return func(app *App) {
		app.kubeContext = context
	}
}

func WithNamespace(namespace string) func(app *App) {
	return func(app *App) {
		app.namespace = namespace
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
		app.container = containerName
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

func WithArgs(args []string) func(app *App) {
	return func(app *App) {
		app.args = args
	}
}

func WithDestPath(destPath string) func(app *App) {
	return func(app *App) {
		app.destPath = destPath
	}
}

func WithRunner(runner string) func(app *App) {
	return func(app *App) {
		app.runner = runner
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
		log.Error().Msg("Error: context, file, and pod name are required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if app.fileType != "auto" && app.fileType != "script" && app.fileType != "binary" {
		log.Error().Msg("Error: type must be 'auto', 'script', or 'binary'")
		os.Exit(1)
	}
}
