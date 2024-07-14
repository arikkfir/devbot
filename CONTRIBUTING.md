# Contributing

Hi there! We're thrilled that you'd like to contribute to this project. Your help is essential for keeping it great.

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.

## Setup

Local development is done by installing the necessary tools, creating a local `kind` cluster, along with an
observability layer that helps debugging the system and gaining insights into what it does.

This can be accomplished by the following commands:

```bash
$ brew install make go kind kubebuilder kubernetes-cli kustomize skaffold yq jq node
$ make setup
$ make create-local-cluster
$ make setup-observability
```

## Developing

The easiest way to get up & running once a local cluster has been created, is using Skaffold to (continuously) deploy
the Helm chart into the cluster. Skaffold will keep the deployed chart updated with any code changes you make
automatically, and can be run from the CLI, or from a JetBrains/VSCode plugin.

Here is how to run the application from the CLI, assuming a cluster has been created:

```bash
$ make dev
```

This will run `skaffold dev` which will package & deploy the Helm chart, and will keep doing so as you make changes to
the code.

## Testing

To run the full tests suite locally, [ensure you have a running cluster](#setup) and that `devbot` is deployed to it
(see [Developing](#developing) above), then run the following:

```bash
$ make e2e
```

## Issues and PRs

If you have suggestions for how this project could be improved, or want to report a bug, open an issue! We'd love all
and any contributions. If you have questions, too, we'd love to hear them.

We'd also love PRs. If you're thinking of a large PR, we advise opening up an issue first to talk about it, though! Look
at the links below if you're not sure how to open a PR.

## Submitting a pull request

1. [Fork](https://github.com/arikkfir/devbot/fork) and clone the repository.
2. Make sure you set up your local development environment as described above.
3. Create a new branch: `git checkout -b my-branch-name`.
4. Make your change, add your feature/bug specific tests, and make sure the entire tests suite passes (see above)
5. Push to your fork, and submit a pull request
6. Pat your self on the back and wait for your pull request to be reviewed and merged.

Here are a few things you can do that will increase the likelihood of your pull request being accepted:

- Write and update tests.
- Keep your changes as focused as possible
  - Smaller PRs make faster & easier reviews, which make faster acceptance & merges
  - Keep each PR focused on one specific change
- Write a [good commit message](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).
- Provide any and all necessary information in the PR description to help reviewers understand the context and impact of
  your change.

Work in Progress pull requests are also welcome to get feedback early on, or if there is something blocked you.

## Resources

- [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
- [Using Pull Requests](https://help.github.com/articles/about-pull-requests/)
- [GitHub Help](https://help.github.com)
