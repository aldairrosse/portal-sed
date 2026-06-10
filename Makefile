.PHONY: test test-unit test-integration test-docker test-docker-up test-docker-down test-all

# Run unit tests only (fast, no DB required)
test-unit:
	cd api && go test ./... -short -v -race -count=1

# Run integration tests (requires PostgreSQL on localhost:5432)
test-integration:
	cd api && go test ./integration/... -v -race -count=1 -timeout=60s

# Run all tests (unit + integration)
test-all:
	cd api && go test ./... -v -race -count=1 -timeout=120s

# Start test containers (PostgreSQL + API)
test-docker-up:
	docker compose -f docker-compose.test.yml up -d --build --wait

# Stop test containers
test-docker-down:
	docker compose -f docker-compose.test.yml down -v

# Run full Docker-based test cycle: up → wait → smoke test → down
test-docker: test-docker-up
	@echo "==> Waiting for API to be ready..."
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		curl -sf http://localhost:8081/health > /dev/null 2>&1 && break; \
		sleep 1; \
	done
	@echo "==> Testing /health endpoint..."
	@curl -sf http://localhost:8081/health && echo ""
	@echo "==> Testing /api/v1/employees endpoint..."
	@curl -sf http://localhost:8081/api/v1/employees | head -c 200 && echo ""
	@echo "==> Testing /api/v1/pillars endpoint..."
	@curl -sf http://localhost:8081/api/v1/pillars | head -c 200 && echo ""
	@echo "==> All smoke tests passed!"
	$(MAKE) test-docker-down

# Run integration tests against Docker test stack
test-docker-integration: test-docker-up
	DATABASE_URL="postgres://sed_test:sed_test@localhost:5433/sed_test?sslmode=disable" \
		cd api && go test ./integration/... -v -race -count=1 -timeout=60s
	$(MAKE) test-docker-down

# Quick lint check
lint:
	cd api && go vet ./...

# Generate Ent code
generate:
	cd api && go generate ./...

# Format code
fmt:
	cd api && gofmt -w .
