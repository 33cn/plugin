
op := "start"

.PHONY: docker-compose help

docker-compose: ## build docker-compose for chain33 run
	@./docker-compose.sh $(op)

docker-compose-down: ## build docker-compose for chain33 run
	@cd temp ;docker-compose down;cd ..

help: ## Display this help screen
	@printf "Help doc:\nUsage: make docker-compose op=[command]\n"
	@printf "[command]\n"
	@printf "[nodegroup]: create super node group if not create \n"
	@printf "[wallet]:    set node wallet private key if not set \n"
	@printf "[miner]:    unlock wallet to miner \n"