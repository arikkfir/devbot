# devbot

> Development bot for GitHub

The `devbot` project is a development environment manager, that helps you manage your development environment, in the
most inclusive interpretation of the term:

* Automatic, on-demand, environment management (e.g. branch-specific environment; feature-specific environment; etc)
* Full integration with existing 3rd-party tools and provides such as GitHub, Slack, etc.
* One-stop dashboard to see everything related to product development in one place

## Contributing

Please see the [contributing guide](.github/CONTRIBUTING.md) for details.

## Status

Alpha. Do not use.

- [ ] Review all calls to `Requeue` - many of those are failures that cannot be recovered from; something like "lastAttemptedCommitSHA" is needed in their place
- [ ] Recreate the unit tests
- [ ] Use slugged branch names as `Environment` and `Deployment` object names
- [ ] Support remote clusters
  - Slugging must be intelligent and avoid conflicts when two different branch names would result in the same slug
- [ ] Refactor conditions
  - All objects
    - `Finalizing`: is `True` if the object is being finalized
    - `FailedToInitialize`: is `True` if object initialization failed
    - `Invalid`: is `True` if object spec is invalid; for things that CRD cannot validate on its own
  - `Repository`
    - `Unauthenticated`: is `True` if authentication to Git provider could not be established
  - `Application`
    - `Stale`: is `True` if an environment is missing or redundant or is stale itself
  - `Environment`
    - `Stale`: is `True` if an `Deployment` is missing or redundant or is stale itself
  - `Deployment`
    - `Cloning`: is `True` if repository is being cloned
    - `Baking`: is `True` if resources manifest is being prepared
    - `Applying`: is `True` if resources manifest is being applied to the target cluster
    - `Stale`: is `True` if last applied commit is not the latest commit in the linked repository
- [ ] Setup CI
  - [ ] Linting
  - [ ] Detect and fail on dead code
  - [ ] Build & publish Docker images
  - [ ] Build & publish `devctl`
