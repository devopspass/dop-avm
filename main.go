package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func main() {
	// Path to the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	// Create a Docker client
	cli, err := client.NewClientWithOpts()
	if err != nil {
		fmt.Println("Error creating Docker client:", err)
		return
	}

	ctx := context.Background()
	cli.NegotiateAPIVersion(ctx)

	// Docker volume mappings
	volumeMappings := []string{
		fmt.Sprintf("%s:/ansible", currentDir),
		"/var/run/docker.sock:/var/run/docker.sock",
	}

	// List of directories to check
	directories := []string{
		filepath.Join(getHomeDir(), ".ssh"),
		filepath.Join(getHomeDir(), ".aws"),
		filepath.Join(getHomeDir(), ".azure"),
		filepath.Join(getHomeDir(), ".ansible"),
	}

	// Ansible home dir
	ansibleDir := filepath.Join(getHomeDir(), ".ansible")
	if _, err := os.Stat(ansibleDir); os.IsNotExist(err) {
		// Create .ansible directory
		if err := os.Mkdir(ansibleDir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", ansibleDir, err)
		} else {
			fmt.Printf("Created directory %s\n", ansibleDir)
			// Append to volume mappings
			volumeMappings = append(volumeMappings, fmt.Sprintf("%s:/%s", ansibleDir, ".ansible"))
		}
	}

	// Check if directories exist and add them as volume mappings
	for _, dir := range directories {
		if _, err := os.Stat(dir); err == nil {
			baseDir := filepath.Base(dir)
			volumeMappings = append(volumeMappings, fmt.Sprintf("%s:/root/%s", dir, baseDir))
		}
	}

	// Docker environment variables
	var envs []string
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "ANSIBLE_") ||
			strings.HasPrefix(env, "MOLECULE_") ||
			strings.HasPrefix(env, "GALAXY_") ||
			strings.HasPrefix(env, "AWS_") {
			envs = append(envs, env)
		}
	}

	// Add GOOGLE_APPLICATION_CREDENTIALS volume if the environment variable is set
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		// Use default path based on the operating system
		switch runtime.GOOS {
		case "windows":
			credentialsPath = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "gcloud", "application_default_credentials.json")
		case "darwin":
			credentialsPath = filepath.Join(getHomeDir(), ".config", "gcloud", "application_default_credentials.json")
		default:
			credentialsPath = filepath.Join(getHomeDir(), ".config", "gcloud", "application_default_credentials.json")
		}
	}

	// Add the credentialsPath to the volume mappings if it's not empty
	if credentialsPath != "" {
		if _, err := os.Stat(credentialsPath); err == nil {
			volumeMappings = append(volumeMappings, fmt.Sprintf("%s:%s", credentialsPath, "/root/.config/gcloud/application_default_credentials.json"))
		}
	}

	// Get the program binary name
	binaryName := filepath.Base(os.Args[0])

	// Create container options
	config := &container.Config{
		Image:        "dop-ansible:latest",
		Cmd:          append([]string{binaryName}, os.Args[1:]...),
		Env:          envs,
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
	}
	hostConfig := &container.HostConfig{
		Binds:      volumeMappings,
		AutoRemove: true,
	}

	// Create the container
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "ansible")
	if err != nil {
		fmt.Println("Error creating Docker container:", err)
		return
	}

	// Start the container
	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		fmt.Println("Error starting Docker container:", err)
		return
	}

	// Stream container logs
	logsOptions := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, logsOptions)
	if err != nil {
		fmt.Println("Error streaming container logs:", err)
		return
	}
	defer out.Close()

	// Print container logs
	_, err = io.Copy(os.Stdout, out)
	if err != nil {
		fmt.Println("Error printing container logs:", err)
		return
	}
}

// Function to get the home directory
func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}
	return home
}
