# Detective Files shortcuts. `make help` lists everything.
# All values overridable: `make api PORT=8081`, `make client DF_API_URL=http://host:8080`.

PORT         ?= 8080
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/detectivefiles?sslmode=disable
REDIS_ADDR   ?= localhost:6379
DF_API_URL   ?= http://localhost:8080

# Screenshot rig: `make shot SCENE=auth SHOT=auth_720 [FLOW=menu] [RES=1920x1080] [ARGS="--suspect=tommy_beck"]`.
SCENE ?= auth
SHOT  ?= auth_720
FLOW  ?=
ARGS  ?=
RES   ?= 1280x720

.PHONY: help services services-down api test vet client check shot shots e2e verify

help: ## Show this list.
	@grep -E '^[a-z-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-13s %s\n", $$1, $$2}'

services: ## Start Postgres + Redis (compose, detached).
	docker compose up -d

services-down: ## Stop Postgres + Redis.
	docker compose down

api: ## Run the Go API on the host (migrations auto-run). Needs `make services`.
	cd backend && PORT=$(PORT) DATABASE_URL="$(DATABASE_URL)" REDIS_ADDR="$(REDIS_ADDR)" go run ./cmd/api

test: ## Backend tests (24: auth, rules, accusations, idempotency, recovery).
	cd backend && go test -count=1 ./...

vet: ## gofmt check + go vet.
	cd backend && gofmt -l . && go vet ./...

client: ## Run the Godot client against DF_API_URL.
	DF_API_URL="$(DF_API_URL)" godot --path game

check: ## Godot headless parse check (errors fail it).
	godot --headless --path game --quit

check-ui: ## Load every screen without backend (parses all scripts, shows empty states).
	for s in startup auth main_menu case_select investigation results profile settings; do \
		$(MAKE) shot SCENE=$$s SHOT=empty_$$s; \
	done

shot: ## One screenshot. Needs API up. Vars: SCENE SHOT FLOW RES.
	@test -n "$(SHOT)" || (echo "SHOT required"; exit 1)
	xvfb-run -a godot --rendering-driver opengl3 --resolution $(RES) --path game \
		res://tools/screenshot.tscn -- --scene=res://scenes/$(SCENE).tscn --shot=$(SHOT) \
		$(if $(FLOW),--flow=$(FLOW)) $(ARGS)

shots: ## All key screens at 1280x720 (startup, auth, menu, cases, desk, suspect, accuse, verdict, profile, settings).
	$(MAKE) shot SCENE=startup SHOT=startup_720
	$(MAKE) shot SCENE=auth SHOT=auth_720
	$(MAKE) shot SCENE=main_menu SHOT=menu_720 FLOW=menu
	$(MAKE) shot SCENE=case_select SHOT=cases_720 FLOW=menu
	$(MAKE) shot SCENE=investigation SHOT=invest_720 FLOW=case
	$(MAKE) shot SCENE=investigation SHOT=suspect_720 FLOW=board ARGS=--suspect=tommy_beck
	$(MAKE) shot SCENE=investigation SHOT=clip_720 FLOW=board ARGS=--clip
	$(MAKE) shot SCENE=investigation SHOT=tunis_720 FLOW=tunis
	$(MAKE) shot SCENE=results SHOT=verdict_720 FLOW=verdict
	$(MAKE) shot SCENE=profile SHOT=profile_720 FLOW=menu
	$(MAKE) shot SCENE=settings SHOT=settings_720 FLOW=menu
	ls -la /tmp/dfshots/

e2e: ## Full manual E2E (register..accuse..relogin). Needs API up.
	python3 scripts/e2e.py

verify: ## Backend tests + Godot parse check.
	$(MAKE) test
	$(MAKE) check
