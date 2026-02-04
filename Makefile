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

# container engine
DOCKER := podman
# default module to test
TEST_MODULE ?= host-basic

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
	@echo '             all    build for linux and windows (amd64)'
	@echo '            test    run tests locally (+coverage)'
	@echo '      build-test    build test binaries (linux and windows, amd64)'
	@echo '     modules-doc    update modules documentation'
	@echo '           clear    remove the build artifacts'
	@echo '        security    run gosec and govulncheck'
	@echo '        analysis    run goweight'
	@echo '         version    print the current version'
	@echo ''
	@echo 'Variables:'
	@echo '            GOOS    target OS'
	@echo '          GOARCH    target architecture'

.PHONY: version all security analysis test build-test clean clear container

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

build-test: $(BIN_PREFIX)-module-testing-amd64-linux $(BIN_PREFIX)-module-testing-amd64-windows.exe

# final binary files
$(BIN_PREFIX)-%: $(SRC_FILES) $(MIGRATION_FILES)
	@mkdir -p $(@D)
	GOARCH=$(call goarch,$*) GOOS=$(call goos,$*) GOARM=$(call goarm,$*) $(BUILD) -o $@ agent/main.go

# binaries for module testing purpose
$(BIN_PREFIX)-module-testing-%: $(MODULE_FILES) 
	@mkdir -p $(@D)
	GOARM=$(call goarm,$*) GOARCH=$(call goarch,$*) GOOS=$(call goos,$*) $(BUILD_TEST) -o $@ $(MODULE)/modules

remote-module-testing-%: module-testing
	ID=$$(head /dev/random|md5sum|head -c 8); \
	$(DOCKER) run -d --rm -it --name "$$ID" $$(echo "$*" | sed -e 's,_,/,g'); \
	$(DOCKER) cp $(BIN_PREFIX)-testing-$(GOARCH)-$(GOOS) "$$ID:/tmp/situation"; \
	$(DOCKER) exec -it "$$ID" sh -c '/tmp/situation -module=$(TEST_MODULE) -test.v' \
	$(DOCKER) rm -f "$$ID"

security: gosec.json govulncheck.json
	@cat $<

gosec.json:
	@gosec -fmt json -exclude-dir dev -quiet ./... | jq > $@

govulncheck.json:
	@govulncheck --json ./... | jq > $@

analysis: goweight.json

goweight.json:
	@goweight --json . | jq > $@

modules-doc: $(MODULE_FILES)
	$(GO) run dev/doc/*.go -d pkg/modules -o docs/modules/

test-modules:
	$(GO) test -v -cover -run 'TestAllModules' ./modules

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

clean: clear

container:
	ko build --local --tarball $@ --tags $(VERSION) --tag-only --base-import-paths