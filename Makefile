.PHONY: generate
generate:
	rm -rf backend/api/v1/zz_* deploy/app/crd/*.yaml
	cd backend && go generate ./...

.PHONY: test
test:
	cd backend && ginkgo run -r --fail-fast

.PHONY: delete-github-test-repositories
delete-github-test-repositories:
	./scripts/delete-all-devbot-repositories.sh
