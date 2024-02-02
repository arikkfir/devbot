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

- [x] Simplify status updates to avoid all those "if" statements
- [ ] Review all calls to `Requeue` - many of those are failures that cannot be recovered from; something like "lastAttemptedCommitSHA" is needed in their place
- [x] Recreate the e2e tests to test drive the system
- [ ] Recreate the unit tests
