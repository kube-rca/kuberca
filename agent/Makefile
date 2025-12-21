SHELL := /bin/bash

UV ?= uv
VENV ?= .venv
APP ?= app.main:app
HOST ?= 0.0.0.0
PORT ?= 8082
IMAGE ?=

.PHONY: venv install lint format test run build

venv:
	@if [ ! -d "$(VENV)" ]; then \
		$(UV) venv; \
	else \
		echo "$(VENV) already exists. Skipping venv creation."; \
	fi

install: venv
	. $(VENV)/bin/activate && $(UV) pip install -e ".[dev]"

lint: install
	. $(VENV)/bin/activate && ruff check .

format: install
	. $(VENV)/bin/activate && ruff format .

test: install
	. $(VENV)/bin/activate && pytest

run: install
	. $(VENV)/bin/activate && uvicorn $(APP) --host $(HOST) --port $(PORT)

build:
	@if [ -z "$(IMAGE)" ]; then \
		echo "IMAGE is required. Example: make build IMAGE=<image>"; \
		exit 1; \
	fi
	docker build -t $(IMAGE) .
