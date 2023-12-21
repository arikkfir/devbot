generate: backend/api/v1/zz*.go deploy/app/crd/*.yaml backend/scripts/generators/api-status-conditions/*
	rm -rfv backend/api/v1/zz_* deploy/app/crd/*.yaml
	cd backend && go generate ./...

.PHONY: test
test:
	cd backend && ginkgo run -r

.PHONY: delete-github-test-repositories
delete-github-test-repositories:
	./scripts/delete-all-devbot-repositories.sh
