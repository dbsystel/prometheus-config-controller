# Prometheus

[Prometheus](https://prometheus.io/), a [Cloud Native Computing Foundation](https://cncf.io/) project, is a systems and service monitoring system. It collects metrics from configured targets at given intervals, evaluates rule expressions, displays the results, and can trigger alerts if some condition is observed to be true.

## Introduction

This chart bootstraps a [Prometheus](https://prometheus.io/) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.3+ with Beta APIs enabled

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm install --name my-release charts/prometheus
```
The command deploys Prometheus on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```console
$ helm delete my-release --purge
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Prometheus chart and their default values.

Parameter | Description | Default
--------- | ----------- | -------
`replicaCount` | The number of pod replicas | `2`
`init.repository` | init container image repository | `busybox`
`init.tag` | init container image tag | `{}`
`init.resources.limits.cpu` | init container resources limits for cpu | `10m`
`init.resources.limits.memory` | init container resources limits for memory | `10Mi`
`init.resources.requests.cpu` | init container resources limits for cpu | `1m`
`init.resources.requests.memory` | init container resources limits for memory | `5Mi`
`prometheus.image.repository` | prometheus container image repository | `prom/prometheus`
`prometheus.image.tag` | prometheus container image tag | `v2.9.2`
`prometheus.configFile` | Pprometheus config file | `/etc/prometheus/prometheus.yml`
`prometheus.storageTsdbPath` | Pprometheus storage path | `/prometheus` 
`prometheus.retention` | Prometheus data retention | `15d`
`prometheus.resources.limits.cpu` | prometheus container resources limits for cpu | `4000m`
`prometheus.resources.limits.memory` | prometheus container resources limits for memory | `8Gi`
`prometheus.resources.requests.cpu` | prometheus container resources limits for cpu | `2000m`
`prometheus.resources.requests.memory` | prometheus container resources limits for memory | `2Gi`
`prometheusController.image.repository` | prometheus-controller container image repository | `dockerregistry/prometheus-config-controller`
`prometheusController.image.tag` | prometheus-controller container image tag | `1.0.0`
`prometheusController.url` | The url to reload prometheus | `http://localhost:9090/-/reload`
`prometheusController.id` | The id to specify prometheus | `0`
`prometheusController.path` | The path to store prometheus configs in prometheus-controller container | `/etc/config`
`prometheusController.template` | The path to find template of prometheus config | `/etc/prometheus/prometheus.tmpl`
`prometheusController.logLevel` | The log-level of prometheus-controller | `info`
`imagePullSecrets` | The name of secret for dockerregistry | `dockerregistry`
`volumeClaimTemplates.name` | The name of Persistent Volume für Prometheus storage | `data`
`volumeClaimTemplates.accessModes` | Prometheus server data Persistent Volume access modes | `[ "ReadWriteOnce" ]`
`volumeClaimTemplates.requests.storage` | Prometheus server data Persistent Volume size | `100Gi`
`ingress.host` | The ingress url | `prometheus.myproject.com` 
`service.port` | The port the prometheus uses | `9090`
`securityContext.fsGroup` | Custom security context for prometheus container | `1000`
`terminationGracePeriodSeconds` | Prometheus Pod termination grace period | `10`

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```console
$ helm install charts/prometheus --name my-release \
    --set enabled=true
```

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
$ helm install charts/prometheus --name my-release -f values.yaml
```

> **Tip**: You can use the default [values.yaml](charts/prometheus/values.yaml)

