package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

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

	app.determineFileType(fileInfo)

	tempPath := app.getDestinationPath()

	defer app.cleanupFile(ctx, tempPath)

	if err := app.copyFileToPod(ctx, file, tempPath); err != nil {
		return err
	}

	return app.runFile(ctx, tempPath)
}

func (app *App) determineFileType(fileInfo os.FileInfo) {
	if app.fileType == "auto" {
		app.fileType = "script"
		if fileInfo.Mode()&0o111 != 0 {
			app.fileType = "binary"
		}
	}
}

func (app *App) getDestinationPath() string {
	if app.destPath == "" {
		app.destPath = "/tmp"
	}
	return fmt.Sprintf("%s/%s", app.destPath, filepath.Base(app.filePath))
}

func (app *App) cleanupFile(ctx context.Context, tempPath string) {
	if err := app.client.DeleteFileFromContainer(ctx, app.pod, app.container, tempPath); err != nil {
		log.Warn().Err(err).Msgf("Failed to delete file %s from pod", tempPath)
	} else {
		log.Debug().Msgf("Deleted file %s from pod", tempPath)
	}
}

func (app *App) copyFileToPod(ctx context.Context, file *os.File, tempPath string) error {
	if err := app.client.CopyFileToContainer(ctx, file, app.pod, app.container, tempPath); err != nil {
		return fmt.Errorf("failed to copy file to pod: %w", err)
	}
	return nil
}

func (app *App) runFile(ctx context.Context, tempPath string) error {
	if app.fileType == "script" {
		return app.executeScript(ctx, tempPath)
	}
	return app.executeBinary(ctx, tempPath)
}

func (app *App) executeScript(ctx context.Context, filePath string) error {
	command := app.buildScriptCommand(filePath)
	return app.executeCommand(ctx, command)
}

func (app *App) buildScriptCommand(filePath string) []string {
	if app.runner != "" {
		return append([]string{app.runner, filePath}, app.args...)
	}

	ext := filepath.Ext(filePath)
	runner := app.inferRunner(ext)
	if runner == "" {
		log.Error().Msgf("Unable to infer runner for file extension: %s", ext)
		return []string{filePath}
	}

	return append([]string{runner, filePath}, app.args...)
}

func (app *App) inferRunner(ext string) string {
	runners := map[string]string{
		".js":  "node",
		".py":  "python",
		".rb":  "ruby",
		".sh":  "sh",
		".php": "php",
	}

	if runner, ok := runners[ext]; ok {
		log.Debug().Msgf("Using %s as runner for %s files", runner, ext)
		return runner
	}

	return ""
}

func (app *App) executeBinary(ctx context.Context, filePath string) error {
	command := append([]string{filePath}, app.args...)
	return app.executeCommand(ctx, command)
}

func (app *App) executeCommand(ctx context.Context, command []string) error {
	log.Debug().Msgf("Running command: %s", strings.Join(command, " "))
	return app.client.RunCommandInPod(ctx, command, app.pod, app.container)
}

