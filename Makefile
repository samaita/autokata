dev:
	@echo "Build & Running"
	@go run main.go
deploy:
	@go build -o autokata
	@docker build -t autokata .