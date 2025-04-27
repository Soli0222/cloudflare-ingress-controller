# Cloudflare Ingress Controller

This project implements a Kubernetes custom controller that automatically injects `status.loadBalancer.ingress.hostname` with `<tunnel>.cfargotunnel.com` for Ingress resources that have `spec.ingressClassName` set to "cloudflare". The `TUNNEL_ID` can be specified via a ConfigMap or environment variable.

## Table of Contents

- [Cloudflare Ingress Controller](#cloudflare-ingress-controller)
  - [Table of Contents](#table-of-contents)
  - [Requirements](#requirements)
  - [Installation](#installation)
  - [Usage](#usage)
  - [Configuration](#configuration)
  - [Development](#development)
  - [License](#license)

## Requirements

- Kubernetes 1.16+
- Go 1.16+
- Helm 3.x

## Installation

To install the Cloudflare Ingress Controller, you can use Helm. First, ensure you have Helm installed and configured to connect to your Kubernetes cluster.

1. Clone the repository:

   ```
   git clone <repository-url>
   cd cloudflare-ingress-controller
   ```

2. Install the Helm chart:

   ```
   helm install cloudflare-ingress-controller ./charts/cloudflare-ingress-controller
   ```

## Usage

Once installed, the controller will start watching for Ingress resources with the `ingressClassName` set to "cloudflare". It will automatically update the `status.loadBalancer.ingress.hostname` field with the specified `TUNNEL_ID`.

## Configuration

You can customize the installation by providing a `values.yaml` file or using `--set` flags during installation. The following parameters can be configured:

- `tunnelID`: The Cloudflare tunnel ID to be used in the hostname.

Example of a `values.yaml` file:

```yaml
tunnelID: your-tunnel-id
```

## Development

To build the Docker image for the custom controller, run:

```
docker build -t cloudflare-ingress-controller .
```

To run the controller locally, you can use:

```
go run ./cmd/controller/main.go --kubeconfig <path-to-kubeconfig> --tunnel-id <your-tunnel-id>
```

## License

This project is licensed under the MIT License. See the LICENSE file for more details.