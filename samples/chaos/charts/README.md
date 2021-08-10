# litmus

![Version: 1.16.0](https://img.shields.io/badge/Version-1.16.0-informational?style=flat-square) ![AppVersion: 1.13.8](https://img.shields.io/badge/AppVersion-1.13.8-informational?style=flat-square)

A Helm chart to install litmus infra components on Kubernetes

**Homepage:** <https://litmuschaos.io>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| ksatchit | karthik.s@mayadata.io |  |
| ispeakc0de | shubham@chaosnative.com |  |

## Source Code

* <https://github.com/litmuschaos/litmus>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| appinfo.appkind | string | `nil` |  |
| appinfo.applabel | string | `nil` |  |
| appinfo.appns | string | `nil` |  |
| customLabels | object | `{}` | Additional labels |
| experiments.disabled | list | `[]` |  |
| fullnameOverride | string | `"litmus"` |  |
| ingress.annotations | object | `{}` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts | list | `[{"host":null,"paths":[]}]` |  kubernetes.io/tls-acme: "true" |
| ingress.tls | list | `[]` |  |
| litmusGO.image.pullPolicy | string | `"Always"` |  |
| litmusGO.image.repository | string | `"litmuschaos/go-runner"` |  |
| litmusGO.image.tag | string | `"1.13.8"` |  |
| nameOverride | string | `"litmus"` |  |
| nodeSelector | object | `{}` |  |
| operator.image.pullPolicy | string | `"Always"` |  |
| operator.image.repository | string | `"litmuschaos/chaos-operator"` |  |
| operator.image.tag | string | `"1.13.8"` |  |
| operatorMode | string | `"standard"` |  |
| operatorName | string | `"chaos-operator"` |  |
| policies | object | `{"monitoring":{"disabled":false}}` |  https://docs.litmuschaos.io/docs/faq-general/#does-litmus-track-any-usage-metrics-on-the-test-clusters |
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| runner.image.repository | string | `"litmuschaos/chaos-runner"` |  |
| runner.image.tag | string | `"1.13.8"` |  |
| service.port | int | `80` |  |
| service.type | string | `"ClusterIP"` |  |
| tolerations | list | `[]` |  |

