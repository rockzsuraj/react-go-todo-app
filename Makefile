.PHONY: build-frontend build-backend deploy rollback status clean

# Docker registry
REGISTRY := ghcr.io
REPO := suraj/react-todo
TAG := $(shell git rev-parse --short HEAD)

FRONTEND_IMAGE := $(REGISTRY)/$(REPO)-frontend:$(TAG)
BACKEND_IMAGE := $(REGISTRY)/$(REPO)-backend:$(TAG)

build-frontend:
	docker build -t $(FRONTEND_IMAGE) \
		--build-arg REACT_APP_API_URL=$(API_URL) \
		--build-arg REACT_APP_SUPABASE_URL=$(SUPABASE_URL) \
		--build-arg REACT_APP_SUPABASE_ANON_KEY=$(SUPABASE_ANON_KEY) \
		apps/web

build-backend:
	docker build -t $(BACKEND_IMAGE) apps/api

push-images: build-frontend build-backend
	docker push $(FRONTEND_IMAGE)
	docker push $(BACKEND_IMAGE)

deploy:
	./scripts/deploy.sh deploy --frontend-image=$(FRONTEND_IMAGE) --backend-image=$(BACKEND_IMAGE)

rollback:
	./scripts/deploy.sh rollback

status:
	./scripts/deploy.sh status

clean:
	docker rmi $(FRONTEND_IMAGE) $(BACKEND_IMAGE) || true

setup-secrets:
	@echo "Creating Kubernetes secrets..."
	@kubectl create secret generic todo-secrets \
		--from-literal=db-host=$(DB_HOST) \
		--from-literal=db-user=$(DB_USER) \
		--from-literal=db-password=$(DB_PASSWORD) \
		--from-literal=db-name=$(DB_NAME) \
		-n todo-app --dry-run=client -o yaml | kubectl apply -f -