test:
	@go test -cover ./... -v
bench:
	@go test -count=1 -run=^$$ -bench=. -benchmem ./... | grep -vE "^(\?|ok|PASS) " ; true
clean:
	@find . -type f -name "*.log" -delete
lint:
	@staticcheck ./...
protoc:
	@protoc internal/proto/metric.proto  --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative