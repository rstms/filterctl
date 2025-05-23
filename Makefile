# go makefile

program != basename $$(pwd)

latest_release != gh release list --json tagName --jq '.[0].tagName' | tr -d v

version != cat VERSION

gitclean = if git status --porcelain | grep '^.*$$'; then echo git status is dirty; false; else echo git status is clean; true; fi

build: fmt 
	fix go build

fmt: go.sum
	fix go fmt . ./...

go.mod:
	go mod init
	go get github.com/rstms/mabctl@$(shell gh release --repo rstms/mabctl list --json tagName --jq '.[0].tagName')
	go get github.com/rstms/rspamd-classes@$(shell gh release --repo rstms/rspamd-classes list --json tagName --jq '.[0].tagName')

go.sum: go.mod
	go mod tidy

install: build
	go install

test: fmt build
	go test -failfast -v . ./...

debug: fmt
	go test -v -failfast -run $(test) . ./...


release:
	@$(gitclean) || { [ -n "$(dirty)" ] && echo "allowing dirty release"; }
	@$(if $(update),gh release delete -y v$(version),)
	gh release create v$(version) --notes "v$(version)"

README.md: cmd/usage.go
	echo "# filterctl\n" >$@
	./filterctl usage | jq -r '.Help|.[]' >>$@

testclean:
	rm -f testdata/*.out
	rm -f testdata/*.err

clean: testclean
	rm -f $(program)
	go clean

sterile: clean
	go clean -r || true
	go clean -cache
	go clean -modcache
	rm -f go.mod go.sum
	rm -f README.md
