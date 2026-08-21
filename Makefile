.PHONY: all web bin test clean run

all: web bin test
	@echo MemGit build complete.

web:
	@echo Building Svelte frontend...
	@cd web && npm install && npm run build

bin: web
	@echo Compiling MemGit server binary...
	@go build -ldflags="-s -w" -o bin/memgit-server.exe ./cmd/memgit-server
	@echo Server binary compiled to bin/memgit-server.exe

test:
	@echo Running tests...
	@go test -v ./...

clean:
	@echo Cleaning artifacts...
	@powershell -Command "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue bin, web/dist"
	@echo Clean complete.

run:
	@powershell -Command "./bin/memgit-server.exe -port 8500"
