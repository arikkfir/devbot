# Contributing

Hi there! We're thrilled that you'd like to contribute to this project. Your help is essential for keeping it great.

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.

## Setup

### Tools

```bash
$ brew install yq jq                                                  # used occasionally by various scripts
$ brew install kubebuilder                                            # might be useful, not strictly required currently
$ brew install kubernetes-cli kustomize                               # will be used by skaffold
$ brew install skaffold datawire/blackbird/telepresence-arm64         # local development tools
$ brew install redis                                                  # for local inspection of redis data
$ brew install go                                                     # for backend development
$ go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0  # used to generate CRDs from controller code
$ brew install node                                                   # for smee-client
$ npm install -g smee-client                                          # used by tests to tunnel webhook requests
```

### Local Kubernetes cluster

```bash
$ brew install kind                                               # local development cluster
$ brew install helm                                               # for installing packages into the cluster
$ helm repo add datawire https://app.getambassador.io             # add the datawire repo (for telepresence)
$ helm repo update                                                # update the repo index
$ kubectl create namespace ambassador                             # prepare namespace for telepresence
$ helm install traffic-manager datawire/telepresence \
    --namespace ambassador \
    --set ambassador-agent.enabled=false
```

## Issues and PRs

If you have suggestions for how this project could be improved, or want to report a bug, open an issue! We'd love all and any contributions. If you have questions, too, we'd love to hear them.

We'd also love PRs. If you're thinking of a large PR, we advise opening up an issue first to talk about it, though! Look at the links below if you're not sure how to open a PR.

## Submitting a pull request

1. [Fork](https://github.com/arikkfir/devbot/fork) and clone the repository.
2. Configure and install the dependencies: `npm install`.
3. Make sure the tests pass on your machine: `npm test`, note: these tests also apply the linter, so there's no need to lint separately.
4. Create a new branch: `git checkout -b my-branch-name`.
5. Make your change, add tests, and make sure the tests still pass.
6. Push to your fork and [submit a pull request](https://github.com/arikkfir/devbot/compare).
7. Pat your self on the back and wait for your pull request to be reviewed and merged.

Here are a few things you can do that will increase the likelihood of your pull request being accepted:

- Write and update tests.
- Keep your changes as focused as possible. If there are multiple changes you would like to make that are not dependent upon each other, consider submitting them as separate pull requests.
- Write a [good commit message](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).

Work in Progress pull requests are also welcome to get feedback early on, or if there is something blocked you.

## Resources

- [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
- [Using Pull Requests](https://help.github.com/articles/about-pull-requests/)
- [GitHub Help](https://help.github.com)
