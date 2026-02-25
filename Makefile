# 
#        - Situation Go Agent Makefile -
# 
#             ╔▄▄▓▓█████████▓▄▄µ.               
#         ;▄▓████████████████████▓▄µ            
#       ╗▓█████▓▀╨""'     '"╙▀▓█████▓▄.         
#     ╗▓████▌╨'    .,╓╓╓╓,.     ╙▀█████▄        
#   ,▓████▀'   ,▄▓██████████▓▄,   '╢████▓≈      
#  ╔▓███▓╨   ╔▓██████▓▓▓▓██████▓φ   "▀████w     
#  ▓███▌"  .╣████▌╨'      '╙▀████▌¡  '╣████,    
# ╫████M  .╣████╨   .╗▄▄╗,   ╙▓███▌,  ╙▓███▌    
# ▓███▌'  ╟████╨   ▄██████▌-  ╙▓███▒   ╢████    
# ▓███▌   ║████   .████████M   ▓███▌   ╟████    
# ▓███▌.  ╟████N   ╝████████▄,╦▓███▒   ╢███▓    
# ╠████Ñ   ╣████N    "╙╙▀█████████▌'  ╔▓███▒    
#  ▀███▓m   ╢████▓▄u.    ,╫██████▌'  ╔▓███▓'    
#  '▀████N   "▀███████████████████▓▄╦▓███▓'     
#    ▀████▓µ    ╙▀▓████████▓▀╨╙▀████████▌'      
#     ╙▀████▓▄¿      ''''      ╓╣█████▓╨        
#       ╙▀██████▓▄▄µ;.  .,╔╗▄▓██████▌╨          
#          ╨▀████████████████████▀╨'            
#             '╙▀▀▀▓███████▓▀▀╙'                      
# 
# 

MODULE     := $(shell go list -m)
VERSION    := 0.20.0
COMMIT     := $(shell git rev-parse HEAD)

# system stuff
GO          := $(shell which go)
GOARCH      ?= $(shell go env|grep GOARCH|awk -F '=' '{print $$2}'|sed -e "s/[']//g")
GOOS        ?= $(shell go env|grep GOOS  |awk -F '=' '{print $$2}'|sed -e "s/[']//g")
CGO_ENABLED ?= 0

# files
SRC_FILES       := $(shell find . -path "*.go" -not -path "./.*")
MODULE_FILES    := $(shell find ./pkg/modules -path "*.go")
MIGRATION_FILES	:= $(shell find ./pkg/store/migrations -path "*.sql")


# Put the version in the config file
GO_LDFLAGS_SET_VERSION := -X "$(MODULE)/agent/config.Version=$(VERSION)"
# Put the conmmit in the config file
GO_LDFLAGS_SET_COMMIT  := -X "$(MODULE)/agent/config.Commit=$(COMMIT)"
# Put the conmmit in the config file
GO_LDFLAGS_SET_MODULE  := -X "$(MODULE)/agent/config.Module=$(MODULE)"
# if cgo link statically
GO_LDFLAGS_STATIC_LINK := -linkmode external -extldflags "-static"
# default ld flags
GO_LDFLAGS_BASE        := $(GO_LDFLAGS_SET_VERSION) $(GO_LDFLAGS_SET_COMMIT) $(GO_LDFLAGS_SET_MODULE)
# Omit the symbol table and debug information.
GO_LDFLAGS_STRIP       := -s
# Omit the DWARF symbol table
GO_LDFLAGS_STRIP_DWARF := -w
# prod flags
GO_LDFLAGS_PROD        := $(GO_LDFLAGS_STRIP) $(GO_LDFLAGS_STRIP_DWARF)
# final flags
GO_LDFLAGS             ?= -ldflags '$(GO_LDFLAGS_BASE) $(GO_LDFLAGS_PROD)'

# build command
BUILD      := CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GO_LDFLAGS) -trimpath
BUILD_TEST := CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(GO_LDFLAGS) -c

# name of the final binary
BIN        := situation
BIN_DIR    := bin
BIN_PREFIX := $(BIN_DIR)/$(BIN)-$(VERSION)

# postgres test database 
PG_DSN     ?= postgresql://postgres:situation@127.0.0.1:5432/postgres?sslmode=disable

# utils
split = $(word $2,$(subst -, ,$(basename $1)))
# Extract GOOS from filename (e.g., amd64-windows.exe -> windows)
goos = $(call split,$1,2)
# Extract GOARCH from filename (armv*l/armhf/armel -> arm)
goarch = $(if $(filter armv%l armhf armel,$(call split,$1,1)),arm,$(call split,$1,1))
# Extract GOARM from filename (armv5l->5, armv6l->6, armv7l->7, armhf->7, armel->5)
goarm = $(strip \
  $(if $(filter armhf,$(call split,$1,1)),7, \
  $(if $(filter armel,$(call split,$1,1)),5, \
  $(patsubst armv%l,%,$(filter armv%l,$(call split,$1,1))))))


.DEFAULT_GOAL := $(BIN_PREFIX)-$(GOARCH)-$(GOOS)
.DEFAULT:
	@echo -e '\033[31mUnknown command "$@"\033[0m'
	@echo 'Usage: make [command] [variable=]...'
	@echo ''
	@echo 'Commands:'
	@echo '             all    build for linux and windows'
	@echo '            test    run tests locally (+coverage)'
	@echo '            docs    update modules documentation'
	@echo '           clear    remove the build artifacts'
	@echo '        security    run gosec and govulncheck'
	@echo '        analysis    run goweight'
	@echo '         version    print the current version'
	@echo ''

.PHONY: version all security analysis test clean clear container

version:
	@echo "$(VERSION)"

go.sum: go.mod 

go.mod:
	$(GO) mod init $(MODULE)
	$(GO) mod tidy

all: $(BIN_PREFIX)-amd64-linux \
	 $(BIN_PREFIX)-arm64-linux \
	 $(BIN_PREFIX)-armv5l-linux \
	 $(BIN_PREFIX)-armv6l-linux \
	 $(BIN_PREFIX)-armv7l-linux \
	 $(BIN_PREFIX)-amd64-windows.exe

# final binary files
$(BIN_PREFIX)-%: $(SRC_FILES) $(MIGRATION_FILES)
	@mkdir -p $(@D)
	GOARCH=$(call goarch,$*) GOOS=$(call goos,$*) GOARM=$(call goarm,$*) $(BUILD) -o $@ agent/main.go

security: gosec.json govulncheck.json
	@cat $<

gosec.json:
	@gosec -fmt json -exclude-dir dev -quiet ./... | jq > $@

govulncheck.json:
	@govulncheck --json ./... | jq > $@

analysis: goweight.json

goweight.json:
	@goweight --json . | jq > $@

docs: modules-doc

modules-doc: $(MODULE_FILES)
	$(GO) run internal/main.go modules-doc
	$(GO) run internal/main.go db-doc

test: .gocoverprofile.html

.gocoverprofile.txt: $(shell find . -path "*_test.go")
	$(GO) test -coverprofile=$@ -covermode=atomic $$(go list -v ./...| grep -v pkg/modules)

.gocoverprofile.html: .gocoverprofile.txt
	$(GO) tool cover -html=$^ -o $@

clear:
	rm -f $(BIN_PREFIX)-*
	rm -f .go*.json
	rm -f .go*.txt
	rm -f .go*.html
	rm -rf sdk/*/*.ts sdk/**/*.py sdk/**/dist sdk/**/*.tgz sdk/**/*.gz sdk/**/*.whl 

clean: clear

container:
	ko build --local --tarball $@ --tags $(VERSION) --tag-only --base-import-paths


# SDK generators (experimental)
# NOTE: a migrated pg instance is required (PG_DSN should be set accordingly)

sdk: sdk/drizzle/situation-sh-drizzle-$(VERSION).tgz \
	 sdk/sqlmodel/dist/situation_sdk-$(VERSION)-py3-none-any.whl

sdk/drizzle/situation-sh-drizzle-$(VERSION).tgz: $(MIGRATION_FILES)
	@mkdir -p sdk/drizzle
	@sed -i 's/"version":[ ]*".*"/"version": "$(VERSION)"/' sdk/drizzle/package.json
	bun run drizzle-kit pull --out sdk/drizzle --url "$(PG_DSN)" --dialect postgresql
	@printf "export * from './schema';\nexport * from './relations';\n" > sdk/drizzle/index.ts
	cd sdk/drizzle && bun run build && bun pm pack

sdk/sqlmodel/dist/situation_sdk-$(VERSION)-py3-none-any.whl: $(MIGRATION_FILES)
	@mkdir -p sdk/sqlmodel
	@sed -i 's/"version"[ ]*[=][ ]*".*"/"version" = "$(VERSION)"/' sdk/sqlmodel/pyproject.toml
	uv run sqlacodegen --generator sqlmodels --outfile sdk/sqlmodel/situation.py "$(PG_DSN)" 
	cd sdk/sqlmodel && uv build -o dist

