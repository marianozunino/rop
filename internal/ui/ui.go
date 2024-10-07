package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

func RunContainerSelection(containers []string) (string, error) {
	if len(containers) == 0 {
		return "", fmt.Errorf("no containers available")
	}

	var choice string
	err := huh.NewSelect[string]().
		Title("Multiple containers detected, please select one:").
		Options(huh.NewOptions(containers...)...).
		Value(&choice).
		Run()
	if err != nil {
		return "", fmt.Errorf("error running selection: %w", err)
	}

	if choice == "" {
		return "", fmt.Errorf("no container selected")
	}

	return choice, nil
}

const (
	commandStyleStr   = "3" // Yellow
	podStyleStr       = "6" // Cyan
	containerStyleStr = "5" // Magenta
)

var (
	commandStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(commandStyleStr))
	podStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color(podStyleStr))
	containerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(containerStyleStr))
)

func ConfirmAction(command, podName, container string) error {
	var confirm bool
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Description(
				fmt.Sprintf("Are you sure you want to execute '%s' on pod '%s' in container '%s'?",
					commandStyle.Render(command),
					podStyle.Render(podName),
					containerStyle.Render(container)),
			),
			huh.NewConfirm().Title("Confirm action?").Affirmative("Yes").Negative("No").Value(&confirm),
		),
	).Run()
	if err != nil {
		return fmt.Errorf("error running confirmation: %w", err)
	}

	if !confirm {
		return fmt.Errorf("action aborted by user")
	}

	return nil
}
