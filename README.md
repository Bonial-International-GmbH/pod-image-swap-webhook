# pod-image-swap-webhook

[![Build Status](https://github.com/Bonial-International-GmbH/pod-image-swap-webhook/actions/workflows/ci.yml/badge.svg)](https://github.com/Bonial-International-GmbH/pod-image-swap-webhook/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Bonial-International-GmbH/pod-image-swap-webhook)](https://goreportcard.com/report/github.com/Bonial-International-GmbH/pod-image-swap-webhook)
![License](https://img.shields.io/github/license/Bonial-International-GmbH/pod-image-swap-webhook)

A mutating webhook that patches Pod container images based on configuration
rules. For example, if you want to transparently proxy image pulls through an
internal registry, this webhook might be for you.

## Preparation

This webhook works best in combination with
[Harbor](https://github.com/goharbor/harbor) and its [proxy
cache](https://goharbor.io/docs/2.1.0/administration/configure-proxy-cache/)
feature.

It is recommended to setup a proxy cache project for every registry for which
you want the webhook to replace images.

Of course this webhook also works without Harbor.

## Deployment

The helm chart provided in this repository can be used to deploy the webhook.

First, add the helm repository:

```sh
helm repo add pod-image-swap-webhook \
  https://bonial-international-gmbh.github.io/pod-image-swap-webhook
```

Create a `values.yaml` and add a `webhookConfig` section with the desired
replacement configuration, for example:

```yaml
---
webhookConfig:
  exclude:
    - prefix: k8s.gcr.io/ingress-nginx/controller
    - regexp: ^quay\.io/.*prometheus.*$
  replace:
    - prefix: quay.io
      replacement: registry.example.org/quay.io
    - prefix: k8s.gcr.io
      replacement: registry.example.org/k8s.gcr.io
    - prefix: docker.io
      replacement: registry.example.org/docker.io
    - regexp: ^.*busybox(:(?P<tag>.+))?$
      replacement: registry.example.org/library/busybox:1.35
```

You can find documentation for all available `webhookConfig` fields in
[`config.sample.yaml`](config.sample.yaml). For more helm configuration options
have a look into the [`values.yaml`
defaults](charts/pod-image-swap-webhook/values.yaml).

Finally use helm to install the webhook:

```sh
helm upgrade pod-image-swap-webhook pod-image-swap-webhook/pod-image-swap-webhook \
  --install --namespace kube-system --values values.yaml
```

## License

The source code of pod-image-swap-webhook is released under the MIT License.
See the bundled LICENSE file for details.
