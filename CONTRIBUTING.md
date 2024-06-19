# Contributing

Hi there! We're thrilled that you'd like to contribute to this project. Your help is essential for keeping it great.

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.

## Setup

### Install required toolchains

```bash
$ brew install go node                                                # language toolchains
$ brew install kind kubebuilder kubernetes-cli kustomize skaffold     # Kubernetes development tooling
$ brew install yq jq                                                  # useful tools often used ad-hoc or by scripts
$ go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0  # used to generate CRDs from controller code
$ npm install -g smee-client                                          # used by tests to tunnel webhook requests
```

### Create a local Kubernetes cluster

```bash
$ ./hack/bin/setup-cluster.sh
```

### Developing

In our experience, the best development experience is to use Skaffold in conjunction with its respective IDE plugin 
(either JetBrains or VSCode usually). This allows for a very fast development cycle, where changes are automatically
reflected in the cluster.

This will also allow for local debugging of the code.

### Testing

To run the full tests suite locally, you can do the following:

```bash
$ skaffold build -q | skaffold deploy --build-artifacts=-
$ go test ./...                   # run tests
$ skaffold delete                 # undeploy from the local cluster
```

## Issues and PRs

If you have suggestions for how this project could be improved, or want to report a bug, open an issue! We'd love all
and any contributions. If you have questions, too, we'd love to hear them.

We'd also love PRs. If you're thinking of a large PR, we advise opening up an issue first to talk about it, though! Look
at the links below if you're not sure how to open a PR.

## Submitting a pull request

1. Make sure you set up your local development environment as described above.
2. [Fork](https://github.com/arikkfir/devbot/fork) and clone the repository.
3. Create a new branch: `git checkout -b my-branch-name`.
4. Make your change, add your feature/bug specific tests, and make sure the entire tests suite passes (see above)
5. Push to your fork, and submit a pull request
6. Pat your self on the back and wait for your pull request to be reviewed and merged.

Here are a few things you can do that will increase the likelihood of your pull request being accepted:

- Write and update tests.
- Keep your changes as focused as possible
  - Break up your change to smaller, separate & decoupled changes - reviewing will be easier & faster
  - Keep each PR focused on one specific change
- Write a [good commit message](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).
- Provide any and all necessary information in the PR description to help reviewers understand the context and impact of
  your change.

Work in Progress pull requests are also welcome to get feedback early on, or if there is something blocked you.

## Resources

- [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
- [Using Pull Requests](https://help.github.com/articles/about-pull-requests/)
- [GitHub Help](https://help.github.com)
