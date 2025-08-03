set dotenv-load := true

CAPNPROTO_IMAGE := "proto-compiler"

check-compiler:
	docker image inspect {{CAPNPROTO_IMAGE}} >/dev/null 2>&1 || docker build -f Dockerfile.tools -t {{CAPNPROTO_IMAGE}} .

build-proto:
  just check-compiler
  docker run --rm \
  -v $(PWD):/app -w /app \
  {{CAPNPROTO_IMAGE}} \
  sh -e ./scripts/build-proto.sh

sqlc-generate:
  sqlc generate

build:
  CGO_ENABLED=0 go build -o dist/service main.go

run:
  go run main.go

dev: sqlc-generate
	find . -name "*.go" | entr -rc just run

compose-up:
  docker compose up -d

compose-down:
  docker compose down -v

publish-image:
  docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/carlosqsilva/rinha-2025:latest --push .
