set quiet := true

# Display help
[private]
help:
  just --list --unsorted

run:
  go run ./cmd/tcplistener/

test:
  go test ./...
