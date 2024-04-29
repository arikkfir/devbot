.PHONY: debug-print-resources
debug-print-resources:
	./hack/bin/debug-deployments.sh

.PHONY: generate
generate:
	rm -vrf backend/api/v1/zz_* deploy/app/crd/*.yaml
	cd backend && go generate ./...

.PHONY: delete-github-test-repositories
delete-github-test-repositories:
	gh repo list devbot-testing --json=nameWithOwner \
      | jq '.[]|.nameWithOwner' -r \
      | sort \
      | xargs -I@ gh repo delete --yes @
#      | xargs -I@ op plugin run -- gh repo delete --yes @

.PHONY: delete-local-cluster
delete-local-cluster:
	kind get clusters | grep -q devbot && kind delete cluster -n devbot

.PHONY: create-local-cluster
create-local-cluster: delete-local-cluster
	kind create cluster -n devbot

.PHONY: ensure-local-cluster
ensure-local-cluster:
	kind get clusters | grep -q devbot || kind create cluster -n devbot

.PHONY: test
test:
	cd backend && go test ./...

.PHONY: tidy
tidy:
	cd backend && go mod tidy

.PHONY: skaffold-dev
dev: ensure-local-cluster
	skaffold dev
