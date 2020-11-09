OTELCOL_BUILDER_VERSION ?= 0.1.3
OTELCOL_BUILDER_DIR ?= ~/bin
OTELCOL_BUILDER ?= $(OTELCOL_BUILDER_DIR)/opentelemetry-collector-builder

build: otelcol-builder
	@$(OTELCOL_BUILDER) --config manifest.yaml

otelcol-builder:
ifeq (, $(shell which opentelemetry-collector-builder))
	@{ \
	set -e ;\
	mkdir -p $(OTELCOL_BUILDER_DIR) ;\
	curl -sLo $(OTELCOL_BUILDER) https://github.com/observatorium/opentelemetry-collector-builder/releases/download/v$(OTELCOL_BUILDER_VERSION)/opentelemetry-collector-builder_$(OTELCOL_BUILDER_VERSION)_linux_x86_64 ;\
	chmod +x $(OTELCOL_BUILDER) ;\
	}
else
OTELCOL_BUILDER=$(shell which opentelemetry-collector-builder)
endif
