package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

const (
	// DefaultAudience is the audience we established works with your AWS setup
	DefaultAudience = "sts.amazonaws.com"

	// MetadataURL is the endpoint to get the ID token
	MetadataURL = "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity"
)

func main() {
	audience := DefaultAudience
	if len(os.Args) > 1 {
		audience = os.Args[1]
	}

	// 1. Determine output file path
	uid := os.Getuid()
	defaultPath := fmt.Sprintf("/run/user/%d/aws_gcp_token", uid)

	outputFile := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
	if outputFile == "" {
		outputFile = defaultPath
	}

	// 2. Check for Linger (Warning only)
	checkLinger()

	// 3. Fetch Token
	token, err := fetchToken(audience)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: ERROR: Error fetching token: %v\n", time.Now().Format(time.RFC3339), err)
		os.Exit(1)
	}

	// 4. Write to file atomically
	if err := atomicWriteFile(outputFile, token); err != nil {
		fmt.Fprintf(os.Stderr, "%s: ERROR: Error writing to file: %v\n", time.Now().Format(time.RFC3339), err)
		os.Exit(1)
	}

	// 5. Output for shell evaluation and logs to stderr
	fmt.Fprintf(os.Stderr, "%s: SUCCESS: Successfully wrote token to %s\n", time.Now().Format(time.RFC3339), outputFile)
	fmt.Printf("export AWS_WEB_IDENTITY_TOKEN_FILE=%s\n", outputFile)
}

func checkLinger() {
	currUser, err := user.Current()
	if err != nil {
		return
	}

	lingerPath := filepath.Join("/var/lib/systemd/linger", currUser.Username)
	if _, err := os.Stat(lingerPath); os.IsNotExist(err) {
		// Only warn if we are on a system that likely uses systemd
		if _, err := os.Stat("/var/lib/systemd"); err == nil {
			fmt.Fprintf(os.Stderr, "%s: WARNING: Linger is not enabled for user %s. Background cron jobs may fail to access /run/user. Run 'loginctl enable-linger %s' to fix.\n",
				time.Now().Format(time.RFC3339), currUser.Username, currUser.Username)
		}
	}
}

func fetchToken(audience string) (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Construct the full URL with query parameters
	req, err := http.NewRequest("GET", MetadataURL, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("audience", audience)
	q.Add("format", "full")
	req.URL.RawQuery = q.Encode()

	// Required header for GCP Metadata Server
	req.Header.Add("Metadata-Flavor", "Google")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("metadata server returned status %d: %s", resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

func atomicWriteFile(filename string, data string) error {
	// Create a temporary file in the same directory
	dir := filepath.Dir(filename)
	// Ensure the directory exists
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// CreateTemp creates files with 0600 permissions by default
	tmpFile, err := os.CreateTemp(dir, "token-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	// Write data
	if _, err := tmpFile.WriteString(data); err != nil {
		tmpFile.Close()
		return err
	}

	// Ensure data is on disk
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	// Atomic rename
	return os.Rename(tmpFile.Name(), filename)
}

