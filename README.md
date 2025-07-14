# Sports Community Management App with Keycloak Authorization

This project demonstrates a sports community management application for soccer associations, using Keycloak for role-based access control
and authorization policies. The application allows multiple organizations (soccer associations) to manage their sports events
securely.

## Overview

The application implements the following structure:

* **Organizations:** Each soccer association represents a separate organization
* **Roles:**
    * **System Admin:** Overall system administrator
    * **Organization Maintainer:** Manages association settings and users
    * **Organization User:** Regular users (parents) within an association
* **Resources:** Sports events managed by each organization
* **Components:**
    * **Keycloak:** Handles authentication and fine-grained authorization
    * **Go Backend Service:** Manages sports events and organization data
    * **Frontend:** User interface for managing / listing events

## Testuser

* Username: `f.roeser+demo@smight.com`
* Password: `Test1234567890!`

## Next Steps



## Documentation

Learn more about the technical aspects of this project through our documentation:

* [Authentication Basics](docs/01-authentication-basics.md) - Understanding OAuth 2.0, PKCE flow, and token validation
* [Authentication Architecture](docs/02-service-setup.md) - Detailed explanation of the authentication and authorization architecture using Keycloak

### Project Structure

The project follows Go best practices for directory structure:

```
/
├── backend/                # Backend service
│   ├── cmd/                # Command-line applications
│   │   └── api/            # Main API server application
│   │       └── main.go     # Entry point for the API server
│   ├── internal/           # Private application code
│   │   ├── handlers/       # HTTP request handlers
│   │   │   ├── events.go   # Event handlers
│   │   │   └── routes.go   # API routes
│   │   ├── models/         # Data models
│   │   │   └── event.go    # Event model
│   │   └── repository/     # Data access layer
│   │       ├── events.go   # Events repository interface
│   │       └── postgres_events.go # Postgres implementation
│   ├── Dockerfile          # Docker build configuration
│   └── go.mod              # Go module definition
├── frontend/               # Frontend application
│   ├── html/               # HTML templates and assets
│   │   ├── css/            # CSS styles
│   │   ├── js/             # JavaScript files
│   │   └── templates/      # Angular templates
│   ├── Dockerfile          # Docker build configuration
│   └── nginx.conf          # Nginx configuration
├── data/                   # Data initialization
│   ├── db/                 # Database scripts
│   └── import/             # Keycloak import files
├── Dockerfile              # Docker build configuration for the API server
├── docker-compose.yml      # Docker Compose configuration
└── Magefile.go             # Mage build tasks
```

#### Go Best Practices

This project follows these Go best practices for directory structure:

1. **cmd/**: Contains the main applications for the project. Each subdirectory is a separate executable.
2. **internal/**: Contains private application code that should not be imported by other projects.
3. **Modular structure**: The backend is organized as a separate Go module with its own go.mod file.
4. **Repository pattern**: Data access is abstracted through repository interfaces.
5. **Absolute imports**: All imports use absolute paths based on the module path, not relative paths.

## Prerequisites

Before you can run this project, make sure you have the following installed:

*   **Go:** Programming language for the backend service ([https://go.dev/dl/](https://go.dev/dl/))
*   **Docker:** Container platform ([https://docs.docker.com/get-docker/](https://docs.docker.com/get-docker/))
*   **Docker Compose:** Tool for defining and running multi-container Docker applications ([https://docs.docker.com/compose/install/](https://docs.docker.com/compose/install/))
*   **Mage:** Go-based build tool ([https://magefile.org/](https://magefile.org/))

## Local Development Setup

1.  **Clone the Repository:**
    ```bash
    git clone <your_repository_url>
    cd <your_repository_directory>
    ```

2.  **Install Mage:**
    ```bash
    go install github.com/magefile/mage@latest
    ```

3.  **Build and Start Services:**
    ```bash
    mage build   # Build all Docker images
    mage start   # Start all services with docker-compose
    ```

    Or simply run:
    ```bash
    mage         # Builds and starts all services
    ```

4.  **Access the Application:**
    * Frontend: http://localhost:80
    * Backend API: http://localhost:8080
    * Keycloak Admin Console: http://localhost:8081 (admin/bad-password)

## Usage

*   **Access the Frontend:**  Open http://localhost:80 in your browser.
*   **Interact with the Backend:**  The frontend will guide you through the OAuth login flow, using Keycloak for authentication.
*   **Inspect Logs:**  Use `mage logs` to view the logs of all services for debugging.

## Available Mage Commands

*   **mage test**: Run all tests (backend and frontend)
*   **mage build**: Build all Docker images
*   **mage start**: Start all services with docker-compose
*   **mage stop**: Stop all services
*   **mage clean**: Remove all built artifacts
*   **mage logs**: Show logs from all services
*   **mage**: Build and start all services (default command)

## Configuration

*   **Docker Compose:**  Customize the deployment by modifying the `docker-compose.yml` file.
*   **Keycloak:**  Configure Keycloak users, realms, and clients through the Keycloak Admin Console (http://localhost:8081 - admin/bad-password).

## Example Use Cases

* Organization maintainers can create and manage sports events
* Parents (users) can view their organization's event schedule
* System administrators can manage all organizations
* Users are restricted to viewing only their organization's data

## Contributing

Feel free to contribute to the project by submitting pull requests.

## License

[Choose a suitable license, e.g., MIT License]
