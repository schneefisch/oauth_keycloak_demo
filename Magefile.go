//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var pidDir = ".pids"

// savePID saves a process PID to a file
func savePID(name string, pid int) error {
	os.MkdirAll(pidDir, 0755)
	return os.WriteFile(filepath.Join(pidDir, name+".pid"),
		[]byte(fmt.Sprintf("%d", pid)), 0644)
}

// readPID reads a PID from a file
func readPID(name string) (int, error) {
	data, err := os.ReadFile(filepath.Join(pidDir, name+".pid"))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

// killProcess kills a process by name
func killProcess(name string) {
	if pid, err := readPID(name); err == nil {
		if p, err := os.FindProcess(pid); err == nil {
			p.Kill()
		}
		os.Remove(filepath.Join(pidDir, name+".pid"))
	}
}

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

// Start starts infrastructure via docker, backend and frontend directly
func Start() error {
	// 1. Start infrastructure (keycloak + postgres)
	fmt.Println("Starting infrastructure (Keycloak + PostgreSQL)...")
	if err := sh("docker-compose up -d"); err != nil {
		return fmt.Errorf("failed to start infrastructure: %w", err)
	}
	fmt.Println("Infrastructure started, waiting for services...")
	time.Sleep(5 * time.Second)

	// 2. Start backend
	fmt.Println("Starting backend...")
	backendCmd := exec.Command("go", "run", "./cmd/api")
	backendCmd.Dir = "backend"
	backendCmd.Env = append(os.Environ(),
		"DB_HOST=localhost",
		"DB_PORT=5432",
		"DB_USER=admin",
		"DB_PASSWORD=admin",
		"DB_NAME=events_demo",
		"KEYCLOAK_URL=http://localhost:8081",
		"SERVER_PORT=8082",
	)
	backendCmd.Stdout = os.Stdout
	backendCmd.Stderr = os.Stderr
	if err := backendCmd.Start(); err != nil {
		return fmt.Errorf("failed to start backend: %w", err)
	}
	savePID("backend", backendCmd.Process.Pid)

	// 3. Start frontend with npx serve
	fmt.Println("Starting frontend...")
	frontendCmd := exec.Command("npx", "serve", "-l", "8080")
	frontendCmd.Dir = "frontend/html"
	frontendCmd.Stdout = os.Stdout
	frontendCmd.Stderr = os.Stderr
	if err := frontendCmd.Start(); err != nil {
		return fmt.Errorf("failed to start frontend: %w", err)
	}
	savePID("frontend", frontendCmd.Process.Pid)

	fmt.Println("\nAll services started:")
	fmt.Println("  Frontend: http://localhost:8080")
	fmt.Println("  Backend:  http://localhost:8082")
	fmt.Println("  Keycloak: http://localhost:8081")
	return nil
}

// Stop stops all services
func Stop() error {
	fmt.Println("Stopping backend...")
	killProcess("backend")

	fmt.Println("Stopping frontend...")
	killProcess("frontend")

	fmt.Println("Stopping infrastructure...")
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
