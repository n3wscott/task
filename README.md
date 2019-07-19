# Task

[![GoDoc](https://godoc.org/github.com/n3wscott/task?status.svg)](https://godoc.org/github.com/n3wscott/task)
[![Go Report Card](https://goreportcard.com/badge/knative/sample-controller)](https://goreportcard.com/report/knative/sample-controller)

Task creates _Addressable_ _PodSpecable_ Jobs that run to completion.

The _Addressable_ aspect gives Task.status.address.url from a Kubernetes Service.

The _PodSpecable_ aspect allows you to provide a pod template in Task.

Task will run once and only once.

## Installing

To install the latest release, 

```shell
kubectl apply -f https://github.com/n3wscott/task/releases/download/v0.2.0/release.yaml
```

To install from master using [ko](github.com/google/ko),

```shell
ko -f ./config
```

---

To learn more about Knative, please visit our
[Knative docs](https://github.com/knative/docs) repository.

If you are interested in contributing, see [CONTRIBUTING.md](./CONTRIBUTING.md)
and [DEVELOPMENT.md](./DEVELOPMENT.md).
