DOCKER_REPO=daruzero
DOCKER_IMG=cfautoupdater-go
DOCKER_TAG=$(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))
docker_image=$(DOCKER_REPO)/$(DOCKER_IMG):$(DOCKER_TAG)
TEST_FLAGS=
DIRS= ./...
TAG_INC=patch

.PHONY: all

go-deps: go-tidy go-vendor

go-tidy:
	@go mod tidy

go-vendor:
	@go mod vendor

go-test:
	@go test $(TEST_FLAGS) $(DIRS)

go-fmt:
	@go fmt $(DIRS)
	@fieldalignment -fix $(DIRS)

docker: docker-build docker-push

docker-build:
	docker buildx build -t $(docker_image) -f build/package/Dockerfile .

docker-push:
	docker push $(docker_image)

tag:
	@./scripts/tag.sh $(TAG_INC)
	git push --tags