.DEFAULT_GOAL := up

# ========================================================================================
# Start everything with local profile.

.PHONE: up
up: build app-up clean

# Stop everything with local profile.

.PHONE: down
down: app-down clean

# Log everything with local profile.

.PHONE: logs
logs: app-logs


# ========================================================================================
# Builds all the go binaries.

.PHONY: build
build:
	@echo building binaries...

	@go build -C app/cmd/gorilla
	
# ========================================================================================
# Clean binaries

.PHONY: clean
clean:
	@echo cleaning binaries

	@go clean all

# ========================================================================================
# Start containers.

.PHONY: app-up
app-up:
	@docker compose -f ./app/compose.yaml --profile local up -d --build

# ========================================================================================
# Stop containers.

.PHONY: app-down
app-down:
	@docker compose -f ./app/compose.yaml --profile local down

# ========================================================================================
# Display logs.

.PHONE: app-logs
app-logs:
	@echo "******************** app LOGS ********************"

	@docker compose -f ./app/compose.yaml --profile local logs

# ========================================================================================