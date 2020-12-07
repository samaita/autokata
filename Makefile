run:
	@echo "Build & Running"
	@go build -o autokata && ./autokata
deploy:
	@go build -o autokata
	@docker build -t autokata .