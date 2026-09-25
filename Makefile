.PHONY: build test vet audit run smoke
build:
	go build -o bin/basket ./cmd/basket
test:
	go test -race ./...
vet:
	go vet ./...
audit:
	python3 tools/audit_parity.py
	python3 tools/generate_openapi.py
	python3 tools/generate_api_reference.py
run:
	go run ./cmd/basket -config config.yaml

smoke:
	python3 tools/smoke_report.py
