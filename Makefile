
run:
	go run ./cmd/card-shoggoths-server

dev:
	air

test:
	go test ./internal/... ./cmd/...

dump-to-clipboard:
	while read fn ; do \
		echo -ne "\n\n##### $$fn\n\n" ; \
		cat "$$fn" ; \
		echo ; \
	done < <( find -name Makefile -or -name go.mod -name '*.go' -or -name '*.html' -or -name '*.js' -or -name '*.css' ) | xclip
