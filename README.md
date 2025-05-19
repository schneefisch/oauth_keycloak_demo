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
    * **Kubernetes:** Deployment and management platform

## Prerequisites

Before you can run this project, make sure you have the following installed:

*   **kubectl:** Kubernetes command-line tool ([https://kubernetes.io/docs/tasks/tools/](https://kubernetes.io/docs/tasks/tools/))
*   **Helm:** Package manager for Kubernetes ([https://helm.sh/docs/intro/install/](https://helm.sh/docs/intro/install/))
*   **Go:** Programming language for the backend service ([https://go.dev/dl/](https://go.dev/dl/))
*   **Minikube or Kind:**  Local Kubernetes cluster (choose one) -  [https://minikube.sigs.k8s.io/docs/](https://minikube.sigs.k8s.io/docs/) or [https://kind.sigs.k8s.io/](https://kind.sigs.k8s.io/)

## Setup

1.  **Clone the Repository:**
    ```bash
    git clone <your_repository_url>
    cd <your_repository_directory>
    ```

2.  **Start Local Kubernetes Cluster:**
    *   **Minikube:**
        ```bash
        minikube start
        ```
    *   **Kind:**
        ```bash
        kind create cluster
        ```

3.  **Add the dependencies:**
    ```bash
    cd charts/keycloak/
    helm dependency update
    ```

4.  **Build Docker Images:**

    Build the Docker images for the backend and frontend:

    ```bash
    # Build backend image
    cd backend
    docker build -t backend-service:latest .
    cd ..

    # Build frontend image
    cd frontend
    docker build -t frontend:latest .
    cd ..
    ```

5.  **Deploy Charts:**

    Navigate to the `charts` directory:

    ```bash
    cd charts
    ```

    Deploy each of the charts. Replace `<your-namespace>` with your desired namespace (e.g., `oauth-demo`):

    ```bash
    helm install keycloak keycloak/keycloak -n <your-namespace> --create-namespace
    helm install backend-service ./backend-service -n <your-namespace> --create-namespace
    helm install frontend ./frontend -n <your-namespace> --create-namespace
    ```

## Usage

*   **Access the Frontend:**  Find the frontend service's ingress or port information from `kubectl get svc -n <your-namespace>`.  Access the frontend URL in your browser.
*   **Interact with the Backend:**  The frontend will guide you through the OAuth login flow, using Keycloak for authentication.
*   **Inspect Logs:**  Use `kubectl logs` to view the logs of the different components for debugging.

## Configuration

*   **Helm Values:**  Customize the deployment of each component by modifying the `values.yaml` files in the `charts` directory.
*   **Keycloak:**  Configure Keycloak users, realms, and clients through the Keycloak Admin Console ([https://localhost:8080](https://localhost:8080) - adapt URL according to your cluster setup).

## Example Use Cases

* Organization maintainers can create and manage training appointments
* Parents (users) can view their organization's training schedule
* System administrators can manage all organizations
* Users are restricted to viewing only their organization's data

## Contributing

Feel free to contribute to the project by submitting pull requests.

## License

[Choose a suitable license, e.g., MIT License]
