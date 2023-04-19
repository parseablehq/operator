<h2 align="center">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/parseablehq/.github/main/images/logo-dark.png">
      <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/parseablehq/.github/main/images/logo.svg">
      <img alt="Parseable Logo" src="https://raw.githubusercontent.com/parseablehq/.github/main/images/logo.svg">
    </picture>
    <br>
    Parseable Kubernetes Operator
</h2>

<div align="center">

[![Docker Pulls](https://img.shields.io/docker/pulls/parseable/parseable?logo=docker&label=Docker%20Pulls)](https://hub.docker.com/r/parseable/parseable)
[![Slack](https://img.shields.io/badge/slack-brightgreen.svg?logo=slack&label=Community&style=flat&color=%2373DC8C&)](https://launchpass.com/parseable)
[![Docs](https://img.shields.io/badge/stable%20docs-parseable.io%2Fdocs-brightgreen?style=flat&color=%2373DC8C&label=Docs)](https://www.parseable.io/docs)
[![Build](https://img.shields.io/github/checks-status/parseablehq/parseable/main?style=flat&color=%2373DC8C&label=Checks)](https://github.com/parseablehq/parseable/actions)

</div>

Parseable is a lightweight, cloud native log observability engine. Written in Rust, Parseable is built for high ingestion rates and low resource consumption. It is compatible with all major log agents and can be configured to collect logs from any source. Read more in [Parseable docs](https://www.parseable.io/docs).

## Parseable Operator

The Parseable Kubernetes operator deploys and manages Parseable instances in a Kubernetes cluster. The operator allows creating multi-tenant Parseable instances.

## Installation

The Parseable operator can be installed using Helm:

```bash
helm repo add parseable https://charts.parseable.io
helm install parseable-operator parseable/operator --create-namespace --namespace parseable-operator
kubectl apply -f https://raw.githubusercontent.com/parseablehq/operator/main/config/samples/parseable-ephemeral.yaml
```

## Attribution

Parseable operator uses [DSOI Spec](https://github.com/datainfrahq/dsoi-spec) and [Operator Runtime](https://github.com/datainfrahq/operator-runtime) project to decouple application logic from the Operator CRD.
