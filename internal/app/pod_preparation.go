package app

import (
	"fmt"
	"os"

	"github.com/marianozunino/rop/internal/ui"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
)

// PreparePodEnvironment handles the preparation steps for pod execution
func (app *App) PreparePodEnvironment() error {
	if err := app.selectExecutionContainer(); err != nil {
		return fmt.Errorf("container selection failed: %w", err)
	}

	return nil
}

func (app *App) selectExecutionContainer() error {
	if app.container != "" {
		log.Debug().Msgf("Using pre-selected container: %s", app.container)
		return nil
	}

	containers := app.pod.Spec.Containers

	if len(containers) == 1 {
		app.container = containers[0].Name
		log.Debug().Msgf("Single container found, using: %s", app.container)
		return nil
	}

	return app.promptForContainer(containers)
}

func (app *App) promptForContainer(containers []corev1.Container) error {
	containerNames := make([]string, len(containers))
	for i, container := range containers {
		containerNames[i] = container.Name
	}

	selectedContainer, err := ui.RunContainerSelection(containerNames)
	if err != nil {
		return fmt.Errorf("error running container selection: %w", err)
	}

	app.container = selectedContainer
	log.Debug().Msgf("Selected container: %s", app.container)
	return nil
}

func (app *App) validateInputFile() error {
	fileInfo, err := os.Stat(app.filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", app.filePath)
	}
	if err != nil {
		return fmt.Errorf("error checking input file: %w", err)
	}

	log.Debug().Msgf("Input file '%s' exists, size: %d bytes", app.filePath, fileInfo.Size())
	return nil
}

