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

MODULE     := github.com/situation-sh/situation
VERSION    := 0.19.1
COMMIT     := $(shell git rev-parse HEAD)

# system stuff
GO        := $(shell which go)
GOARCH    ?= $(shell go env|grep GOARCH|awk -F '=' '{print $$2}'|sed -e "s/[']//g")
GOOS      ?= $(shell go env|grep GOOS  |awk -F '=' '{print $$2}'|sed -e "s/[']//g")

# files
SRC_FILES    := $(shell find . -path "*.go" -not -path "./.*")
MODULE_FILES := $(shell find ./modules -path "*.go")


# Put the version in the config file
GO_LDFLAGS_SET_VERSION := -X "$(MODULE)/config.Version=$(VERSION)"
# Put the conmmit in the config file
GO_LDFLAGS_SET_COMMIT  := -X "$(MODULE)/config.Commit=$(COMMIT)"
# if cgo link statically
GO_LDFLAGS_STATIC_LINK := -linkmode external -extldflags "-static"
# default ld flags
GO_LDFLAGS_BASE        := $(GO_LDFLAGS_SET_VERSION) $(GO_LDFLAGS_SET_COMMIT)
# Omit the symbol table and debug information.
GO_LDFLAGS_STRIP       := -s
# Omit the DWARF symbol table
GO_LDFLAGS_STRIP_DWARF := -w
# prod flags
GO_LDFLAGS_PROD        := $(GO_LDFLAGS_STRIP) $(GO_LDFLAGS_STRIP_DWARF)
# final flags
GO_LDFLAGS             ?= -ldflags '$(GO_LDFLAGS_BASE) $(GO_LDFLAGS_PROD)'

# build command
BUILD      := CGO_ENABLED=0 $(GO) build $(GO_LDFLAGS)
BUILD_TEST := CGO_ENABLED=0 $(GO) test $(GO_LDFLAGS) -c

# name of the final binary
BIN        := situation
BIN_DIR    := bin
BIN_PREFIX := $(BIN_DIR)/$(BIN)-$(VERSION)

# container engine
DOCKER := podman
# default module to test
TEST_MODULE ?= host-basic

# utils
dash-split = $(word $2,$(subst -, ,$1))


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

all: $(BIN_PREFIX)-$(GOARCH)-linux $(BIN_PREFIX)-$(GOARCH)-windows.exe

build-test: $(BIN_PREFIX)-module-testing-amd64-linux $(BIN_PREFIX)-module-testing-amd64-windows.exe

# final binary files
$(BIN_PREFIX)-%: $(SRC_FILES)
	@mkdir -p $(@D)
	GOARCH=$(call dash-split,$(basename $*),1) GOOS=$(call dash-split,$(basename $*),2) $(BUILD) -o $@ main.go

# binaries for module testing purpose
$(BIN_PREFIX)-module-testing-%: $(MODULE_FILES)
	@mkdir -p $(@D)
	GOARCH=$(call dash-split,$(basename $*),1) GOOS=$(call dash-split,$(basename $*),2) $(BUILD_TEST) -o $@ $(MODULE)/modules

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

analysis: .goweight.json

goweight.json:
	@goweight --json . | jq > $@

modules-doc: $(MODULE_FILES)
	$(GO) run dev/doc/*.go -d modules -o docs/modules/

test-modules:
	$(GO) test -v -cover -run 'TestAllModules' ./modules

test: .gocoverprofile.html

.gocoverprofile.txt: $(shell find . -path "*_test.go")
	$(GO) test -coverprofile=$@ -covermode=atomic $$(go list -v ./...| grep -v modules)

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