.PHONY: all get build standalone icon locales generate_locales clean

TAGS ?= gtk_3_18

all: icon locales get build

get:
	go get -tags $(TAGS) . ./standalone

build:
	go build -tags $(TAGS) -ldflags "-X main.version=`git describe --tags`"

test:
	go test -tags $(TAGS) ./...

standalone:
	go build -tags "$(TAGS) standalone" -ldflags "-X main.version=`git describe --tags`"

build_win:
	powershell -Command '$$version=git describe --tags; go build -tags "standalone" -ldflags "-H windowsgui -X main.version=$$version"'

clean:
	make -C icon clean
	rm bitmask-systray

icon:
	make -C icon

get_deps:
	 sudo apt install libgtk-3-dev libappindicator3-dev golang pkg-config


LANGS ?= $(foreach path,$(wildcard locales/*/messages.gotext.json),$(patsubst locales/%/messages.gotext.json,%,$(path)))
empty :=
space := $(empty) $(empty)
lang_list := $(subst $(space),,$(foreach lang,$(LANGS),$(lang),))

locales: catalog.go

generate_locales: $(foreach lang,$(LANGS),locales/$(lang)/out.gotext.json)

locales/%/out.gotext.json: systray.go notificator.go
	gotext update -lang=$*

catalog.go: $(foreach lang,$(LANGS),locales/$(lang)/messages.gotext.json)
	gotext update -lang=$(lang_list) -out catalog.go
