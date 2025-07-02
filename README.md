# Sports Community Management App with Keycloak Authorization

This project demonstrates a sports community management application for soccer associations, using Keycloak for role-based access control
and authorization policies. The application allows multiple organizations (soccer associations) to manage their training appointments
securely.

## Overview

The application implements the following structure:

* **Organizations:** Each soccer association represents a separate organization
* **Roles:**
    * **System Admin:** Overall system administrator
    * **Organization Maintainer:** Manages association settings and users
    * **Organization User:** Regular users (parents) within an association
* **Resources:** Training appointments managed by each organization
* **Components:**
    * **Keycloak:** Handles authentication and fine-grained authorization
    * **Go Backend Service:** Manages training appointments and organization data
    * **Frontend:** User interface for managing / listing appointments

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

* Organization maintainers can create and manage training appointments
* Parents (users) can view their organization's training schedule
* System administrators can manage all organizations
* Users are restricted to viewing only their organization's data

## Contributing

Feel free to contribute to the project by submitting pull requests.

## License

[Choose a suitable license, e.g., MIT License]
