.DEFAULT_GOAL := help

include .env
export  $(shell sed 's/=.*//' .env)

buildx:
	go build -o build/main ./src

up:
	clear && go run ./src

run:
	./build/main

help: ## Info Makefile Tags ~ Mode General
	@printf "\033[31m%-22s %-59s %s\033[0m\n" "Target" " Help"; \
	printf "\033[31m%-22s %-59s %s\033[0m\n"  "------" " ----"; \
	grep -hE '^\S+:.*## .*$$' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' | sort | awk 'BEGIN {FS = ":"}; {printf "\033[32m%-22s\033[0m %-58s \033[34m%s\033[0m\n", $$1, $$2, $$3}'
