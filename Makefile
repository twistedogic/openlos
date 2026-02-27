BIN := openlos

.PHONY: build clean clean_sql test generate generate_sql install

build: clean generate generate_sql
	go build -o $(BIN) .

clean:
	rm -f $(BIN)
	rm -rf assets/opencode

clean_sql:
	rm -rf db/

test:
	go test ./... -timeout 120s

generate:
	go generate ./assets/

generate_sql: clean_sql
	sqlc generate

install: build
	$(BIN) install --dir $(or $(DIR),.)
