test:
	@go test -cover ./... -v
bench:
	@go test -count=1 -run=^$$ -bench=. -benchmem ./... | grep -vE "^(\?|ok|PASS) " ; true
cleaning:
	@find . -type f -name "*.log" -delete