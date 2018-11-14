test:
	go test --race ./...

cover:
	rm -f *.out
	go test -v ./... -coverprofile=coverage.out && go tool cover -func=coverage.out
	go tool cover -html=coverage.out
	rm -f *.out

.PHONY: test cover
