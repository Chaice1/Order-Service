

LOCAL_BIN := $(CURDIR)/bin
GOOSE_BUILD := $(LOCAL_BIN)/goose
BUF_BUILD := $(LOCAL_BIN)/buf

.bin-deps: export GOBIN := $(LOCAL_BIN)
.bin-deps:
	$(info Installing binary dependencies...)
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest


.buf-generate:
	$(info run buf generate...)
	set "PATH=$(LOCAL_BIN);%PATH%" && $(BUF_BUILD) generate

generate: .buf-generate
DB_DSN := "postgres://postgres:Ivbln173@localhost:5432/user_db?sslmode=disable"

migrate-up:
	$(GOOSE_BUILD) -dir migrations postgres $(DB_DSN) up 
migrate-down:
	$(GOOSE_BUILD) -dir migrations postgres $(DB_DSN) down 
migrate-status:
	$(GOOSE_BUILD) -dir migrations postgres $(DB_DSN) status 

.tidy:
	go mod tidy

.PHONY: .bin-deps

# что сделал другую ветку типо user-service-internal и мб + main во ттакое что-то 