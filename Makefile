BACKEND_DIR=backend
FIRMWARE_DIR=firmware
REPORT_DIR=reports

all: help

backend/%:
	$(MAKE) -C $(BACKEND_DIR) $*

firmware/%:
	$(MAKE) -C $(FIRMWARE_DIR) $*

report/%:
	$(MAKE) -C $(REPORT_DIR) $*

help:
	@echo "Command List:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'