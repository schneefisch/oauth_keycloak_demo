# OAuth Demo Project with Keycloak and Go Backend

This project provides a basic demonstration of an OAuth Authorization flow using Keycloak as an Identity Provider and a Go backend service. It's designed for local Kubernetes development and testing.

## Overview

The demo consists of the following components:

*   **Keycloak:** The Identity Provider, responsible for user authentication and authorization.
*   **Go Backend Service:** A simple Go application that requires authentication via OAuth.
*   **Frontend:** A minimal HTML/JavaScript frontend that interacts with the Go backend.
*   **Kubernetes:** The platform for deploying and managing the application components.

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

4.  **Deploy Charts:**

    Navigate to the `charts` directory:

    ```bash
    cd ..
    ```

    Deploy each of the charts. Replace `<your-namespace>` with your desired namespace (e.g., `oauth-demo`):

    ```bash
    helm install keycloak keycloak/keycloak -n <your-namespace> --create-namespace
    helm install frontend ./frontend -n <your-namespace> --create-namespace
    helm install backend-service ./backend-service -n <your-namespace> --create-namespace
    helm install database ./database -n <your-namespace> --create-namespace
    ```

## Usage

*   **Access the Frontend:**  Find the frontend service's ingress or port information from `kubectl get svc -n <your-namespace>`.  Access the frontend URL in your browser.
*   **Interact with the Backend:**  The frontend will guide you through the OAuth login flow, using Keycloak for authentication.
*   **Inspect Logs:**  Use `kubectl logs` to view the logs of the different components for debugging.

## Configuration

*   **Helm Values:**  Customize the deployment of each component by modifying the `values.yaml` files in the `charts` directory.
*   **Keycloak:**  Configure Keycloak users, realms, and clients through the Keycloak Admin Console ([https://localhost:8080](https://localhost:8080) - adapt URL according to your cluster setup).

## Contributing

Feel free to contribute to the project by submitting pull requests.

## License

[Choose a suitable license, e.g., MIT License]