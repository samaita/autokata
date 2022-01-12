dev:
	@echo "Build & Running"
	@air
deploy:
	@go build -o autokata
	@docker build -t autokata .