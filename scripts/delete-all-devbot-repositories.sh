#!/usr/bin/env bash

gh repo list devbot-testing --json=nameWithOwner \
  | jq '.[]|.nameWithOwner' -r \
  | sort \
  | xargs -I@ op plugin run -- gh repo delete --yes @
