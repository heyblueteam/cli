.PHONY: build test vet schema schema-check docs

# The sync targets read from sibling checkouts of the monorepo (../api, ../app).
BLUE_API ?= ../api
BLUE_APP ?= ../app

build:
	go build -o blue ./cmd/blue

test:
	go test ./...

vet:
	go vet ./...

# Refresh the vendored GraphQL schema that `blue api schema` prints.
schema:
	go run ./tools/sync-schema -api $(BLUE_API)

# Fail if the vendored schema has drifted, without rewriting it.
schema-check:
	go run ./tools/sync-schema -api $(BLUE_API) -check

# Refresh the embedded API docs that `blue docs` serves.
docs:
	go run ./tools/sync-docs -source $(BLUE_APP)/src/content/api
