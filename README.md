# devbot

> Development bot for GitHub

## Developer Setup

```bash
# go runtime for backend code
$ brew install go

# controller-gen is used to generate CRDs from controller code
$ go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0

# "smee" is used by end-to-end tests to tunnel webhook requests
$ brew install node
$ npm install -g smee-client
```

## User Setup

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