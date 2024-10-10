/*
Copyright © 2024 Mariano Zunino <marianoz@posteo.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/marianozunino/rop/internal/logger"
	"github.com/minio/selfupdate"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func NewUpdateCmd() *cobra.Command {
	cfg := &config{}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update rop to the latest version",
		Run: func(cmd *cobra.Command, args []string) {
			logger.ConfigureLogger(cfg.verbose)
			if err := runSelfUpdate(http.DefaultClient, afero.NewOsFs(), os.Args[0]); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return updateCmd
}

func init() {
	rootCmd.AddCommand(NewUpdateCmd())
}

func getAssetName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch os {
	case "darwin":
		os = "Darwin"
	case "linux":
		os = "Linux"
	}

	switch arch {
	case "amd64":
		arch = "x86_64"
	case "386":
		arch = "i386"

	}

	return fmt.Sprintf("rop_%s_%s.tar.gz", os, arch)
}

func runSelfUpdate(httpClient *http.Client, fs afero.Fs, executablePath string) error {
	log.Info().Msg("Checking for updates...")

	releaseURL := "https://github.com/marianozunino/rop/releases/latest"

	resp, err := httpClient.Get(releaseURL)
	if err != nil {
		return fmt.Errorf("error checking for updates: %v", err)
	}
	defer resp.Body.Close()

	latestVersionStr := filepath.Base(resp.Request.URL.Path)
	latestVersionStr = strings.TrimPrefix(latestVersionStr, "v")
	currentVersionStr := strings.TrimPrefix(Version, "v")

	latestVersion, err := semver.NewVersion(latestVersionStr)
	if err != nil {
		return fmt.Errorf("error parsing latest version: %v", err)
	}

	currentVersion, err := semver.NewVersion(currentVersionStr)
	if err != nil {
		return fmt.Errorf("error parsing current version: %v", err)
	}

	if !latestVersion.GreaterThan(currentVersion) {
		log.Info().Msg("Current version is the latest")
		return nil
	}

	log.Info().Msgf("New version available: %s (current: %s)", latestVersion, currentVersion)

	assetName := getAssetName()
	downloadURL := fmt.Sprintf("https://github.com/marianozunino/rop/releases/download/v%s/%s", latestVersion, assetName)

	log.Debug().Msgf("Downloading %s from %s", assetName, downloadURL)
	resp, err = httpClient.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("error downloading update: %v", err)
	}
	defer resp.Body.Close()

	// Extract the tar.gz file
	tmpDir, err := os.MkdirTemp("", "rop-update")
	if err != nil {
		return fmt.Errorf("error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	log.Info().Msg("Extracting update package...")
	if err := extractTarGz(resp.Body, tmpDir); err != nil {
		return fmt.Errorf("error extracting update: %v", err)
	}

	// Locate the new binary in the extracted files
	newBinaryPath := filepath.Join(tmpDir, "rop")

	// Apply the update
	newBinary, err := os.Open(newBinaryPath)
	if err != nil {
		return fmt.Errorf("error opening new binary: %v", err)
	}
	defer newBinary.Close()

	log.Info().Msg("Applying update...")
	err = selfupdate.Apply(newBinary, selfupdate.Options{})
	if err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return fmt.Errorf("failed to rollback from bad update: %v", rerr)
		}
		return fmt.Errorf("error updating binary: %v", err)
	}

	log.Info().Msgf("Successfully updated to version %s", latestVersion)
	return nil
}

// extractTarGz extracts a tar.gz file to a specified directory

func extractTarGz(r io.Reader, dst string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("error creating gzip reader: %v", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("error reading tar header: %v", err)
		}

		target := filepath.Join(dst, header.Name)

		// Ensure the directory exists
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("error creating directory %s: %v", target, err)
			}
			continue
		}

		// Create the necessary directories
		if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
			return fmt.Errorf("error creating directory %s: %v", filepath.Dir(target), err)
		}

		// Create a new file
		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
		if err != nil {
			return fmt.Errorf("error creating file %s: %v", target, err)
		}

		// Write the file content
		if _, err := io.Copy(outFile, tr); err != nil {
			return fmt.Errorf("error writing to file %s: %v", target, err)
		}
		outFile.Close()
	}
	return nil
}

