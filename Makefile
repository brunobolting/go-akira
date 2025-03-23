.PHONY: help
help: ## print make targets
	@grep -E '^[a-zA-Z_/-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: air/install
air/install: ## Installs the air build reload system using 'go install'
	@go install github.com/air-verse/air@latest

.PHONY: templ/install
templ/install: ## Installs the templ Templating system for Go using 'go install'
	@go install github.com/a-h/templ/cmd/templ@latest

.PHONY: tailwind/install
tailwind/install: ## Installs the tailwindcss cli
	@curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
	@chmod +x tailwindcss-linux-x64
	@mv tailwindcss-linux-x64 tailwindcss

.PHONY: tailwind/watch
tailwind/watch: ## compile tailwindcss and watch for changes
	@./tailwindcss -i ./static/css/custom.css -o ./static/css/style.css --watch

.PHONY: tailwind/build
tailwind/build: ## one-time compile tailwindcss styles
	@./tailwindcss -i ./static/css/custom.css -o ./static/css/style.css

.PHONY: daisyui/install
daisyui/install: ## Installs the daisyui tailwindcss plugin
	@curl -sLO https://github.com/saadeghi/daisyui/releases/latest/download/daisyui.js
	@mv daisyui.js ./static/plugin/daisyui.js

.PHONY: run
run: ## go run the project
	go run ./cmd/app/main.go

.PHONY: tests
tests: ## run tests
	@go test -timeout 30s -coverprofile=/tmp/coverage -failfast ./...

.PHONY: build
build: ## compile tailwindcss and templ files and build the project
	@./tailwindcss -i ./static/css/custom.css -o ./static/css/style.css --minify
	@templ generate
	@go build -o ./tmp/app ./cmd/app/main.go

.PHONY: air/watch
air/watch: ## build and watch the project with air
	@go build -o ./tmp/app ./cmd/app/main.go && air

.PHONY: templ/build
templ/build: ## generate templ files
	@templ generate

.PHONY: templ/watch
templ/watch: ## generate templ files and watch for changes
	@templ generate --watch

.PHONY: watch
watch: ## build and watch the project and tailwindcss
	@./tailwindcss -i ./static/css/custom.css -o ./static/css/style.css --watch & \
	go build -o ./tmp/app ./cmd/app/main.go && air & \
	wait

.PHONY: db/create
db/create: ## create sqlite database
	@touch db/app.db

.PHONY: db/reset
db/reset: ## reset sqlite database
	@rm db/app.db
	@touch db/app.db
	@GOOSE_DRIVER=sqlite3 GOOSE_DBSTRING=db/app.db goose -dir=./db/migrations up

.PHONY: migration/install
migration/install: ## install goose migration tool
	@go install github.com/pressly/goose/v3/cmd/goose@latest

.PHONY: migration/create
migration/create: ## create a new migration | make migration/create name=migration_name
	@GOOSE_DRIVER=sqlite3 GOOSE_DBSTRING=db/app.db goose -s -dir=./db/migrations create $(name) sql

.PHONY: migration/up
migration/up: ## run all up migrations
	@GOOSE_DRIVER=sqlite3 GOOSE_DBSTRING=db/app.db goose -dir=./db/migrations up

.PHONY: migration/status
migration/status: ## show migration status
	@GOOSE_DRIVER=sqlite3 GOOSE_DBSTRING=db/app.db goose -dir=./db/migrations status

