// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Default target to run when none is specified
var Default = All

// Test runs all tests in the project
func Test() error {
	fmt.Println("Running backend tests...")
	if err := sh("cd backend && go test ./..."); err != nil {
		return err
	}

	fmt.Println("Running frontend tests...")
	// Currently no frontend tests, but we can add them later
	fmt.Println("No frontend tests available")

	return nil
}

// Build builds all services
func Build() error {
	fmt.Println("Building backend...")
	if err := sh("cd backend && docker build -t backend-service:latest ."); err != nil {
		return err
	}

	fmt.Println("Building frontend...")
	if err := sh("cd frontend && docker build -t frontend:latest ."); err != nil {
		return err
	}

	return nil
}

// Start starts all services using docker-compose
func Start() error {
	fmt.Println("Starting all services with docker-compose...")
	return sh("docker-compose up -d")
}

// Stop stops all services
func Stop() error {
	fmt.Println("Stopping all services...")
	return sh("docker-compose down")
}

// All builds and starts all services
func All() error {
	if err := Build(); err != nil {
		return err
	}
	return Start()
}

// Clean removes all built artifacts
func Clean() error {
	fmt.Println("Cleaning up...")
	return sh("docker-compose down -v --rmi local")
}

// Logs shows logs from all services
func Logs() error {
	return sh("docker-compose logs -f")
}

// helper function to run shell commands
func sh(cmd string) error {
	c := exec.Command("sh", "-c", cmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}