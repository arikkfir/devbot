# devbot

> Development bot for GitHub

## Developer Setup

```bash
$ go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0
```

## Setup

### Sources

Define an application source object, similar to this:

```yaml
#file: noinspection KubernetesUnknownResourcesInspection
apiVersion: devbot.app/v1alpha1
kind: GitSource
metadata:
  name: myapp
  namespace: provisioning
spec:
  url: git@github.com:owner/repo.git
  interval: 5m
  auth:
    token:
      secret:
        name: github
        namespace: provisioning # optional; defaults to same namespace
        key: token
```

You can define multiple sources. The following sources are planned:

* GitSource

### 