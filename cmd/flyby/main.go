package main

import (
	"fmt"
	"os"
	"os/exec"

	"flyby/internal/tui"
)

const version = "0.1.0"

func main() {
	// Check if user wants version info
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("FlyBy v%s\n", version)
			fmt.Println("A Terminal UI for Concourse CI")
			os.Exit(0)
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		default:
			fmt.Printf("Unknown option: %s\n", os.Args[1])
			fmt.Println("Use --help for usage information")
			os.Exit(1)
		}
	}

	// Check if fly CLI is available
	if !checkFlyAvailable() {
		fmt.Println("Error: fly CLI not found in PATH")
		fmt.Println("Please install the Concourse fly CLI and ensure it's in your PATH")
		fmt.Println("Download from: https://concourse-ci.org/download.html")
		os.Exit(1)
	}

	app := tui.NewApp()
	if err := app.Run(); err != nil {
		fmt.Printf("Error running FlyBy: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf("FlyBy v%s - Terminal UI for Concourse CI\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  flyby              Start the Terminal UI")
	fmt.Println("  flyby --version    Show version information")
	fmt.Println("  flyby --help       Show this help message")
	fmt.Println("")
	fmt.Println("Features:")
	fmt.Println("  • Manage Concourse targets and teams")
	fmt.Println("  • Browse and manage pipelines")
	fmt.Println("  • Trigger jobs and check resources")
	fmt.Println("  • View build history and status")
	fmt.Println("")
	fmt.Println("Requirements:")
	fmt.Println("  • fly CLI installed and available in PATH")
	fmt.Println("  • Configured Concourse targets in ~/.flyrc")
	fmt.Println("")
	fmt.Println("Navigation:")
	fmt.Println("  • Use arrow keys or j/k to navigate")
	fmt.Println("  • Press Enter to select items")
	fmt.Println("  • Press Esc to go back")
	fmt.Println("  • Press q to quit")
}

func checkFlyAvailable() bool {
	_, err := exec.LookPath("fly")
	return err == nil
}