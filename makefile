.PHONY: all vet sec staticcheck test build

IMAGE_NAME=goiam
TAG=latest
PORT=8080

all: vet sec staticcheck test build

vet:            ; go vet ./...
sec:            ; gosec -exclude-dir=test ./...
#staticcheck:    ; staticcheck -config=.staticcheck.conf ./...
test:           ; go test ./test/...
build:          ; go build -o bin/goiam ./cmd

podman-build:
	podman build -t $(IMAGE_NAME):$(TAG) .

podman-run:
	podman run --rm -p $(PORT):$(PORT) \
		--name $(IMAGE_NAME)-dev \
		$(IMAGE_NAME):$(TAG)

# Start Minikube with Docker (most stable on macOS)
k8s-start:
	minikube start --driver=docker

# Sets shell to use Minikube's Docker daemon
k8s-env:
	eval $$(minikube -p minikube docker-env)

# Build image in the K8S cluster
docker-build: k8s-env
	docker build -t $(IMAGE_NAME):$(TAG) .

# Apply K8S resources (forces recreation of pods)
k8s-deploy: docker-build
	kubectl delete pod -l app=goiam
	kubectl apply -f k8s/

# Open NodePort service in browser
k8s-open:
	minikube service goiam-service

# Tear down all resources
k8s-clean:
	kubectl delete -f k8s/

k8s-logs:
	kubectl logs -f deployment/goiam -c goiam

k8s-shell:
	kubectl exec -it deployment/goiam -c goiam -- /bin/sh