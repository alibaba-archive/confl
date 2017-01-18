test:
	go test --race ./vault
	go test --race ./etcd

cover:
	rm -f *.coverprofile
	go test -race -coverprofile=vault.coverprofile ./vault
	go test -race -coverprofile=etcd.coverprofile ./etcd
	gover
	go tool cover -html=gover.coverprofile
	rm -f *.coverprofile

.PHONY: test cover
