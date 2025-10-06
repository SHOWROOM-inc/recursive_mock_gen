# Makefile

genmock:
	go run cmd/recursive_mock_gen/main.go --output testing/mocks --input .

test:
	go test ./...
