.PHONY: all vet sec staticcheck test build swagger

IMAGE_NAME=goiam
IMAGE_AUTHUI=goiam-authui
TAG=latest
PORT=8080

all: swagger vet sec staticcheck test build

vet:            ; go vet ./...
sec:            ; gosec -exclude-dir=test ./...
#staticcheck:    ; staticcheck -config=.staticcheck.conf ./...
test:           ; go test -short -timeout 30000ms ./...
test-all:           ; go test -timeout 30000ms ./...
build:          ; go build -o bin/goiam ./cmd

docker:
	docker context use orbstack
	docker build -t $(IMAGE_NAME):$(TAG) .

docker-authui:
	docker context use orbstack
	docker build -t $(IMAGE_AUTHUI):$(TAG) ./auth-ui

docker-all: docker docker-authui

docker-run:
	docker run --rm -p $(PORT):$(PORT) \
		--name $(IMAGE_NAME)-dev \
		$(IMAGE_NAME):$(TAG)

helm-deploy:
	helm upgrade --install goam ./helm/goam && kubectl rollout restart deployment/goam && kubectl rollout restart deployment/goam-authui && kubectl rollout status deployment/goam && kubectl rollout status deployment/goam-authui

# Needs: go install github.com/swaggo/swag/cmd/swag@latest
# Note: Uses $(HOME)/go/bin/swag (add to PATH or use full path)
swagger:
	# Generate the swagger documentation
	$$(go env GOPATH)/bin/swag init -g cmd/main.go -o internal/web/swagger-ui --parseInternal