include .env
-include go-service.mk
-include migrations.mk

.PHONY: update-makefile

update-makefile:
	@echo "Updating Makefile..."
	@curl -sSL https://raw.githubusercontent.com/Sokol111/ecommerce-infrastructure/master/makefiles/go-service.mk -o go-service.mk
	@echo "Makefile updated!"
