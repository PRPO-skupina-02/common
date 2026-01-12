test:
	go test ./...

swagger-clients:
	swagger generate client -f ../auth/api/docs/swagger.json -A auth -t ./clients/auth