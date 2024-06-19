#-[ SETUP ]-------------------------------------------------------------------------------------------------------------
GINKGO_FLAGS := -r \
				--procs=1 \
				--seed=0 \
				--fail-fast \
				--fail-on-empty \
				--require-suite \
				--poll-progress-after=5m \
				--poll-progress-interval=1m \
				--timeout=30m \
				--grace-period=5m
ifeq ($(CI),true)
    GINKGO_FLAGS := $(GINKGO_FLAGS) --github-output
endif
ifdef E2E_VERBOSITY # Set E2E_VERBOSITY to "-v" or "-vv" for extra output
    GINKGO_FLAGS := $(GINKGO_FLAGS) $(E2E_VERBOSITY)
else
	GINKGO_FLAGS := $(GINKGO_FLAGS) --succinct
endif
#-----------------------------------------------------------------------------------------------------------------------

deploy/otel-collector-gcp/service-account-key.json:
	gcloud iam service-accounts keys create ./deploy/otel-collector-gcp/service-account-key.json --iam-account=otel-collector@arikkfir.iam.gserviceaccount.com

.PHONY: setup
setup:
	go mod download -x
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.15.0
	go install github.com/onsi/ginkgo/v2/ginkgo@v2.19.0

.PHONY: create-local-cluster
create-local-cluster:
	kind create cluster --name devbot --wait "1m"

.PHONY: setup-observability
setup-observability:
	./scripts/setup-observability.sh

.PHONY: delete-local-cluster
delete-local-cluster:
	kind delete cluster --name devbot || true

#generate: api/v1/zz_generated.application.conditions.go api/v1/zz_generated.constants.go api/v1/zz_generated.deepcopy.go api/v1/zz_generated.deployment.conditions.go api/v1/zz_generated.environment.conditions.go api/v1/zz_generated.repository.conditions.go deploy/chart/crds/devbot.kfirs.com_applications.yaml deploy/chart/crds/devbot.kfirs.com_deployments.yaml deploy/chart/crds/devbot.kfirs.com_environments.yaml deploy/chart/crds/devbot.kfirs.com_repositories.yaml
.PHONY: generate
generate:
	go generate ./...

.PHONY: deploy
deploy:
	skaffold build -q | skaffold deploy --load-images --build-artifacts -

.PHONY: undeploy
undeploy:
	skaffold delete
	kubectl delete --all --all-namespaces deployments.devbot.kfirs.com || true
	kubectl delete --all --all-namespaces environments.devbot.kfirs.com || true
	kubectl delete --all --all-namespaces applications.devbot.kfirs.com || true
	kubectl delete --all --all-namespaces repositories.devbot.kfirs.com || true
	kubectl get namespaces -oname | grep -v -E 'default|devbot|kube|local' | sort | xargs -I@ kubectl delete @  || true
	kubectl delete namespace devbot || true
	kubectl delete crd repositories.devbot.kfirs.com applications.devbot.kfirs.com deployments.devbot.kfirs.com environments.devbot.kfirs.com || true
	gh repo list devbot-testing --json=nameWithOwner \
      | jq '.[]|.nameWithOwner' -r \
      | sort \
      | xargs -I@ gh repo delete --yes @

.PHONY: dev
dev:
	skaffold dev

.PHONY: e2e
e2e:
	DEVBOT_PREFIX=local-devbot ginkgo run $(GINKGO_FLAGS)

.PHONY: e2e-watch
e2e-watch:
	ginkgo watch $(GINKGO_FLAGS)

.PHONY: e2e-unfocus
e2e-unfocus:
	ginkgo unfocus
