package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
	outputFile := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
	if outputFile == "" {
		outputFile = "/tmp/gcp_oidc_token"
	}

	// 2. Fetch Token
	token, err := fetchToken(audience)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: ERROR: Error fetching token: %v\n", time.Now().Format(time.RFC3339), err)
		os.Exit(1)
	}

	// 3. Write to file atomically
	if err := atomicWriteFile(outputFile, token); err != nil {
		fmt.Fprintf(os.Stderr, "%s: ERROR: Error writing to file: %v\n", time.Now().Format(time.RFC3339), err)
		os.Exit(1)
	}

	fmt.Printf("%s: SUCCESS: Successfully wrote token to %s\n", time.Now().Format(time.RFC3339), outputFile)
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
	// If the file is in /tmp, ensure the directory exists (it should), but generic safety:
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist", dir)
	}

	tmpFile, err := os.CreateTemp(dir, "token-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name()) // Clean up if something goes wrong before rename

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
