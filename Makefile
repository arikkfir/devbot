.PHONY: generate
generate:
	rm -vrf api/v1/zz_* deploy/chart/crds/*.yaml
	go generate ./...

.PHONY: delete-e2e-leftovers
delete-e2e-leftovers:
	kubectl get ns -oname | grep -v -E 'default|devbot|kube|local' | sort | xargs -I@ kubectl delete @
	gh repo list devbot-testing --json=nameWithOwner \
      | jq '.[]|.nameWithOwner' -r \
      | sort \
      | xargs -I@ gh repo delete --yes @

.PHONY: delete-local-cluster
delete-local-cluster:
	kind get clusters | grep -q devbot && kind delete cluster -n devbot

.PHONY: create-local-cluster
create-local-cluster: delete-local-cluster
	kind create cluster -n devbot

.PHONY: ensure-local-cluster
ensure-local-cluster:
	kind get clusters | grep -q devbot || kind create cluster -n devbot

.PHONY: skaffold-dev
dev: ensure-local-cluster
	skaffold dev
