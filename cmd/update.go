package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	update "github.com/fynelabs/selfupdate"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update rop to the latest version",
	Long:  `Update rop to the latest version either using go install or by downloading the binary from GitHub.`,
	Run:   runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) {
	fmt.Println("Checking for updates...")

	// Check if the binary was installed using go install
	if installedWithGo() {
		fmt.Println("Updating using go install...")
		if err := updateWithGoInstall(); err != nil {
			fmt.Printf("Error updating with go install: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Updating by downloading the latest release...")
		if err := updateFromGitHub(); err != nil {
			fmt.Printf("Error updating from GitHub: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Update completed successfully!")
}

func installedWithGo() bool {
	// Check if the binary is in GOPATH
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}
	binPath := filepath.Join(gopath, "bin", "rop")
	fmt.Printf("Checking if %s exists...\n", binPath)
	_, err := os.Stat(binPath)
	return err == nil
}

func updateWithGoInstall() error {
	cmd := exec.Command("go", "install", "github.com/marianozunino/rop@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func updateFromGitHub() error {
	// Construct the URL for the latest release
	// Platform with first letter capitalized
	platform := strings.ToUpper(runtime.GOOS[:1]) + runtime.GOOS[1:]

	arch := ""
	switch runtime.GOARCH {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "arm64"
	case "386":
		arch = "i386"
	default:
		return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	url := fmt.Sprintf("https://github.com/marianozunino/rop/releases/latest/download/rop_%s_%s.tar.gz", platform, arch)
	fmt.Printf("URL: %s\n", url)

	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading update: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create a temporary file to store the downloaded content
	tmpfile, err := os.CreateTemp("", "rop_update_*.tar.gz")
	if err != nil {
		return fmt.Errorf("can't create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write the body to file
	_, err = io.Copy(tmpfile, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing to temporary file: %v", err)
	}

	// Extract the binary from the tar.gz file
	binary, err := extractBinary(tmpfile.Name())
	if err != nil {
		return fmt.Errorf("error extracting binary: %v", err)
	}
	fmt.Printf("Extracted binary, size: %d bytes\n", len(binary))

	// Apply the update
	err = update.Apply(bytes.NewReader(binary), update.Options{})
	if err != nil {
		return fmt.Errorf("error applying update: %v", err)
	}

	return nil
}

func extractBinary(archivePath string) ([]byte, error) {
	// Open the tar.gz file
	f, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// Find the binary in the archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if strings.HasSuffix(header.Name, "rop") {
			// Read the entire contents of the file
			binary, err := io.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			return binary, nil
		}
	}

	return nil, fmt.Errorf("binary not found in the archive")
}
