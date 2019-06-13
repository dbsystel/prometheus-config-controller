# Config Controller for Prometheus

This Controller is based on the [Grafana Operator](https://github.com/tsloughter/grafana-operator). The Config Controller should be run within [Kubernetes](https://github.com/kubernetes/kubernetes) as a sidecar with [Prometheus](https://github.com/prometheus/prometheus).

It watches for new/updated/deleted *ConfigMaps* and if they define the specified annotations as `true` it will save each resource from ConfigMap to local storage and reload Prometheus. This requires Prometheus 2.x.

## Annotations

Currently it supports three resources:


**1. Rule**

`prometheus.net/rule` with values: `"true"` or `"false"`

**2. Job**

`prometheus.net/job` with values: `"true"` or `"false"`

**3. Config**

`prometheus.net/config` with values: `"true"` or `"false"`

`prometheus.net/key` with values: `string`

Prometheus will start with a provided minimal dummy config, which is definetly valid. Then the Config Controller will load the *ConfigMap* with annotation `prometheus.net/config: true`, which includes the global Prometheus configuration. The Config Controller will merge this configuration with the dummy configuration and only if this configuration is valid, Prometheus will be reloaded. For each Prometheus Setup there should be only one *ConfigMap* with annotation `prometheus.net/config: true`. If you want to run e.g. two Prometheus instances (replicas = 2), then the two Prometheus instances will load the same ConfigMap and have the exact same config. To prevent other "nonadmin" users from misusing of `prometheus.net/config`, a `key` is used. If and only if the `key` in *ConfigMap* matches the `key` in args of the Prometheus Controller, the Controller will use the *ConfigMap*.

(**Id**)

`prometheus.net/id` with values: `"0"` ... `"n"`

In case of multiple Prometheus *setups* in same Kubernetes Cluster all the ConfigMaps have to be mapped to the right Prometheus setup.
So each *ConfigMap* can be additionaly annotated with the `prometheus.net/id` (if not, the default `id` will be `"0"`)

You can run e.g. two Prometheus instances (replicas = 2) with id=0 and for an another setup in same cluster with two Prometheus instances with id=1, and so on.

**Note**

Mentioned `"true"` values can be also specified with: `"1", "t", "T", "true", "TRUE", "True"`

Mentioned `"false"` values can be also specified with: `"0", "f", "F", "false", "FALSE", "False"`

**ConfigMap examples can be found [here](configmap-examples).**

## Usage
```
--run-outside-cluster # Uses ~/.kube/config rather than in cluster configuration
--reloadUrl # Sets the URL to reload Prometheus
--configPath # Sets the path to use to store rules and config files
--configTemplate # Sets the location of template of the Prometheus config
--id # Sets the ID, so the Controller knows which ConfigMaps should be watched
--key # Sets the key, so the Controller can recognize the template of config in ConfigMap
```

## Development
### Build
```
go build -v -i -o ./bin/prometheus-config-controller ./cmd # on Linux
GOOS=linux CGO_ENABLED=0 go build -v -i -o ./bin/prometheus-config-controller ./cmd # on macOS/Windows
```
To build a docker image out of it, look at provided [Dockerfile](Dockerfile) example.


## Deployment
Our preferred way to install prometheus/prometheus-config-controller is [Helm](https://helm.sh/). See example installation at our [Helm directory](helm) within this repo.
