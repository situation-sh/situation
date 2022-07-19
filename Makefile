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
VERSION    := 0.9.0
COMMIT     := $(shell git rev-parse HEAD)

# system stuff
GO        := $(shell which go)
GOARCH    ?= $(shell go env|grep GOARCH|awk -F '=' '{print $$2}'|sed -e 's/"//g')
GOOS      ?= $(shell go env|grep GOOS  |awk -F '=' '{print $$2}'|sed -e 's/"//g')

# Put the version in the config file
GO_LDFLAGS_SET_VERSION := -X "$(MODULE)/config.Version=$(VERSION)"
# Put the conmmit in the config file
GO_LDFLAGS_SET_COMMIT  := -X "$(MODULE)/config.Commit=$(COMMIT)"
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
BUILD := CGO_ENABLED=0 $(GO) build $(GO_LDFLAGS)

# name of the final binary
BIN     := situation
BIN_DIR := bin

# other tools
GOSEC := $(shell which gosec)

.PHONY: build-linux build-windows

default: build-$(GOOS)-$(GOARCH)

print-version:
	@echo "$(VERSION)"

module:
	rm -f go.mod go.sum
	$(GO) mod init $(MODULE)
	$(GO) mod tidy

all: build-linux-amd64 build-windows-amd64

build-linux-%:
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=$* $(BUILD) -o $(BIN_DIR)/$(BIN)-$(VERSION)-$*-linux main.go 

build-windows-%:
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=$* $(BUILD) -o $(BIN_DIR)/$(BIN)-$(VERSION)-$*-windows.exe main.go 

security:
	$(GOSEC) -fmt json ./...

weight:
	goweight .

docs-module-status:
	@outfile=docs/modules/index.md; 																														\
	rm -f $$outfile; 																															\
	header=$$(cat docs/modules/arp.md|grep -m 20 -e '^|'|sed -e 's,-,,g' -e 's, *| *,|,g'|tail -n +3|awk -F '|' '{print $$2}'|tr '\n' '|');		\
	echo '<div id="modules" markdown>' >> $$outfile;																							\
	echo "|Name|$$header" >> $$outfile;																											\
	bar=$$(echo "$$header"|sed -e 's,[a-zA-Z\ ]*|,---|,g'); 																					\
	echo "|---|$$bar" >> $$outfile;																												\
	for file in docs/modules/*.md; do 																											\
		link=$$(basename $$file);																												\
		name=$$(basename -s .md $$file|sed -e 's,_,-,g');																						\
		if [[ "$$name" == "index" ]]; then continue; fi;																						\
		csv=$$(cat $$file|grep -m 20 -e '^|'|sed -e 's,-,,g' -e 's, *| *,|,g'|tail -n +3|awk -F '|' '{print $$3}'|tr '\n' '|'); 				\
		echo "|[$$name]($$link)|$$csv" >> $$outfile; 																							\
	done;      																																	\
	echo '</div>' >> $$outfile;		 

test:
	$(GO) test -v -coverprofile=coverage.txt -covermode=atomic ./...