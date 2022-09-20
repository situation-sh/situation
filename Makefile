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
VERSION    := 0.12.0
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


# utils
dash-split = $(word $2,$(subst -, ,$1))


.DEFAULT_GOAL := bin/$(BIN)-$(VERSION)-$(GOARCH)-$(GOOS)

.DEFAULT:
	@echo -e '\033[31mUnknown command "$@"\033[0m'
	@echo 'Usage: make [command] [variable=]...'
	@echo ''
	@echo 'Commands:'
	@echo '             all    build for linux and windows (amd64)'
	@echo '           clear    remove the build artifacts'
	@echo '        security    run gosec and govulncheck'
	@echo '        analysis    run goweight'
	@echo '         version    print the current version'
	@echo ''
	@echo 'Variables:'
	@echo '            GOOS    target OS'
	@echo '          GOARCH    target architecture'

version:
	@echo "$(VERSION)"

go.sum: go.mod 

go.mod:
	$(GO) mod init $(MODULE)
	$(GO) mod tidy

all: bin/$(BIN)-$(VERSION)-amd64-linux bin/$(BIN)-$(VERSION)-amd64-windows.exe

$(BIN_DIR)/$(BIN)-$(VERSION)-%: $(shell find . -path "*.go")
	@mkdir -p $(@D)
	GOARCH=$(call dash-split,$(basename $*),1) GOOS=$(call dash-split,$(basename $*),2) $(BUILD) -o $@ main.go

security: .gosec.json .govulncheck.json

.gosec.json:
	@gosec -fmt json ./... | jq > $@

.govulncheck.json:
	@govulncheck --json ./... | jq > $@

analysis: .goweight.json

.goweight.json:
	@goweight --json . | jq > $@



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

test: .goprofile.txt

.goprofile.txt:
	$(GO) test -v -coverprofile=$@ -covermode=atomic ./...


clear:
	rm -f $(BIN_DIR)/$(BIN)-$(VERSION)-*
	rm -f .go*.json
	rm -f .go*.txt