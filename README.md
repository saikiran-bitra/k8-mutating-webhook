# k8-mutating-webhook

This project implements a Kubernetes mutating webhook that adds an `initContainer` to pod specifications. Its primary purpose is to check for the presence of Oracle Java in the images being deployed, preventing the use of Oracle images.

## Table of Contents

1. [Overview](#overview)
2. [Components](#components)
   - [Mutating Webhook Configuration](#mutating-webhook-configuration)
   - [Service YAML](#service-yaml)
   - [Secrets YAML](#secrets-yaml)
   - [Webhook Server Deployment](#webhook-server-deployment)
   - [Dockerfile](#dockerfile)
   - [Webhook Server Code](#webhook-server-code)
3. [Usage](#usage)
4. [Requirements](#requirements)
5. [Contributing](#contributing)

## Overview

The mutating webhook is triggered whenever a pod is created in a namespace labeled with `webhook-poc`. If Oracle Java is detected in the image, the pod creation will be blocked.

## Components

### Mutating Webhook Configuration (`manifest_yamls/webhook.yaml`)
- Sets up a mutating webhook configuration on the cluster.
- Triggered on pod creation in namespaces labeled `webhook-poc`.
- Ensures that pod creation fails if the webhook server cannot be reached (`failurePolicy` is set to `Fail`).
- Requires an updated CA bundle for secure communication with the webhook server. A self-signed certificate can be created for this purpose.

### Service YAML (`manifest_yamls/service-webhook-server.yaml`)
- Defines a service listening on port 443 that routes traffic to the webhook server pod.

### Secrets YAML (`manifest_yamls/secret-webhook-cert-key.yaml`)
- Creates a secret with two entries: one for the certificate (with CN and SAN as `pod-mutator-service.webhook-poc.svc`) and one for the private key.
- Both entries are mounted to the webhook server pod for secure communication.

### Webhook Server Deployment (`manifest_yamls/deployment-webhook-server.yaml`)
- Launches the webhook server pod, which hosts an HTTP server that listens on port 8443.
- The server handles POST requests on the `/mutate` path.
- The admission controller invokes the webhook and sends a POST request with pod information.
- A patch is created to add an `initContainer` that checks for Java make. The initContainer:
  - Exits with `0` if no Oracle Java is found (allowing the main container to start).
  - Exits with `1` if Oracle Java is found (preventing the main container from starting).

### Dockerfile
- Builds the image for the webhook server (`webhook-server.go`).
- Includes commands for configuring private repositories and authentication.
- Ensure to pass the build argument for the token during the Docker build process.

### Webhook Server Code (`webhook-server.go`)
- Contains the source code for the HTTP server that processes POST calls on the `/mutate` resource.
- Returns a patch to add the `initContainer` to the pod spec.

## Usage

To deploy the mutating webhook:

1. Apply the YAML manifests in the following order:
   ```bash
   kubectl apply -f manifest_yamls/secret-webhook-cert-key.yaml
   kubectl apply -f manifest_yamls/service-webhook-server.yaml
   kubectl apply -f manifest_yamls/deployment-webhook-server.yaml
   kubectl apply -f manifest_yamls/webhook.yaml
2. Label your desired namespace:
   ```bash
   kubectl label namespace <your-namespace> webhook-poc=true

## Requirements
Kubernetes cluster
kubectl command-line tool
Docker for building the webhook server image

## Contributing
Contributions are welcome! Please fork the repository and submit a pull request with your changes.

