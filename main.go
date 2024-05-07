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
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

const VERSION = "1.0.0"

func runContainer() {
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

	imageName := os.Getenv("DOP_AVM_IMAGE_NAME")
	if imageName == "" {
		imageName = "devopspass/ansible:latest"
	}

	// Check if the Docker image exists locally
	_, _, err = cli.ImageInspectWithRaw(ctx, imageName)
	if client.IsErrNotFound(err) {
		// Image does not exist locally, pull it
		fmt.Printf("Image %s not found locally, pulling...\n", imageName)
		pullResp, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
		if err != nil {
			fmt.Println("Error pulling Docker image:", err)
			return
		}
		defer pullResp.Close()

		// Stream the pull response if needed
		io.Copy(os.Stdout, pullResp)
	}

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

	// SSH agent
	switch runtime.GOOS {
	case "windows", "darwin":
		volumeMappings = append(volumeMappings, "/run/host-services/ssh-auth.sock:/run/host-services/ssh-auth.sock")
		envs = append(envs, "SSH_AUTH_SOCK=/run/host-services/ssh-auth.sock")
	default:
		if os.Getenv("SSH_AUTH_SOCK") != "" {
			volumeMappings = append(volumeMappings, fmt.Sprintf("%s:%s", os.Getenv("SSH_AUTH_SOCK"), "/tmp/ssh-auth.sock"))
			envs = append(envs, "SSH_AUTH_SOCK=/tmp/ssh-auth.sock")
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
	if runtime.GOOS == "windows" {
		binaryName = strings.Replace(binaryName, ".exe", "", -1)
	}

	// Create container options
	config := &container.Config{
		Image:        imageName,
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
		// Check if the error is due to Docker not running
		if client.IsErrConnectionFailed(err) {
			fmt.Println("ERROR: DOP Ansible Version Manager relies on Docker, so please ensure that it's installed and running.")
			fmt.Println("Error creating Docker container:", err)
		} else {
			fmt.Println("Error creating Docker container:", err)
		}
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

func printHelp() {
	fmt.Printf("DevOps Pass AI - Ansible Version Manager v%s", VERSION)
	fmt.Println("Run 'dop-avm setup' to setup Ansible binaries.")
}

func main() {
	currentBinaryPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting current binary path: %v", err)
		return
	}

	if strings.HasSuffix(currentBinaryPath, "dop-avm") || strings.HasSuffix(currentBinaryPath, "dop-avm.exe") {
		if len(os.Args) > 1 && os.Args[1] == "setup" {
			copyBinaryToNames()
		} else {
			printHelp()
		}
	} else {
		runContainer()
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

func copyBinaryToNames() error {
	// Get the path to the current binary
	currentBinaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting current binary path: %v", err)
	}

	// Determine the target file names
	targetNames := []string{
		"ansible",
		"ansible-playbook",
		"ansible-galaxy",
		"ansible-vault",
		"ansible-doc",
		"ansible-config",
		"ansible-console",
		"ansible-inventory",
		"ansible-adhoc",
		"ansible-lint",
		"molecule",
	}

	// Append ".exe" for Windows
	if runtime.GOOS == "windows" {
		for i, name := range targetNames {
			targetNames[i] = name + ".exe"
		}
	}

	// Copy the binary to each target name
	for _, name := range targetNames {
		targetPath := filepath.Join(".", name)
		if err := copyFile(currentBinaryPath, targetPath); err != nil {
			return fmt.Errorf("error copying binary to %s: %v", targetPath, err)
		}
		fmt.Printf("Binary copied to %s\n", targetPath)
		// Set executable permission for macOS and Linux
		if runtime.GOOS != "windows" {
			if err := os.Chmod(targetPath, 0755); err != nil {
				return fmt.Errorf("error setting executable permission for %s: %v", targetPath, err)
			}
			fmt.Printf("Executable permission set for %s\n", targetPath)
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
