#!/usr/bin/env bash

# TODO: replace "op plugin run -- gh" with "gh" once 1password fixes the issue with the gh plugin

gh repo list devbot-testing --json=nameWithOwner \
  | jq '.[]|.nameWithOwner' -r \
  | sort \
  | xargs -I@ op plugin run -- gh repo delete --yes @
