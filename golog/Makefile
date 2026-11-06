# GoLog build. `make` builds ./golog for this computer; `make release` builds for all platforms.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X golog/internal/server.Version=$(VERSION)

.PHONY: all web build test e2e guide-shots notices dev release clean

all: build

web/node_modules: web/package.json
	cd web && npm install --no-fund --no-audit
	touch web/node_modules

web: web/node_modules
	cd web && npm run build

build: web
	go build -trimpath -ldflags '$(LDFLAGS)' -o golog .

test: web/node_modules
	go vet ./...
	go test ./...
	cd web && npm run check && npm test

# Clicks through the app in headless Firefox (FIREFOX=path to use another; SHOTS=dir for screenshots).
e2e: build
	cd web && npm run e2e

# Retakes the pictures in the user guide (web/public/guide/); check them, then build again.
guide-shots: build
	cd web && node e2e/guide-shots.mjs

# Rewrites web/public/notices.txt (GoLog's licence and those of the libraries built into it)
# from the current dependencies. Run after changing dependencies and commit the result.
notices: web/node_modules
	cd web && node scripts/notices.mjs

# Live-reloading frontend on http://localhost:5173 (API from a running ./golog -no-browser).
dev: web/node_modules
	cd web && npm run dev

# Release packages in dist/: one archive per platform with the program and a short README,
# the licence and third-party notices, plus SHA256SUMS.txt. The version number is in the VERSION file.
RELEASE := $(shell cat VERSION)
RELEASE_LDFLAGS := -s -w -X golog/internal/server.Version=$(RELEASE)
PLATFORMS := linux/amd64 linux/arm64 windows/amd64 darwin/arm64 darwin/amd64

release: notices
	$(MAKE) web
	rm -rf dist && mkdir -p dist
	@set -e; for p in $(PLATFORMS); do \
	  os=$${p%/*}; arch=$${p#*/}; name=$$os; [ $$os = darwin ] && name=macos; \
	  dir=golog-$(RELEASE)-$$name-$$arch; mkdir -p dist/$$dir; \
	  exe=golog; [ $$os = windows ] && exe=golog.exe; \
	  echo "building $$dir"; \
	  CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -trimpath -ldflags '$(RELEASE_LDFLAGS)' -o dist/$$dir/$$exe .; \
	  sed 's/{{VERSION}}/$(RELEASE)/' packaging/README.txt > dist/$$dir/README.txt; \
	  cp LICENSE dist/$$dir/LICENSE.txt; cp web/public/notices.txt dist/$$dir/NOTICES.txt; \
	  if [ $$os = windows ]; then sed -i 's/$$/\r/' dist/$$dir/*.txt; fi; \
	  if [ $$os = linux ]; then tar -C dist -czf dist/$$dir.tar.gz $$dir; else (cd dist && zip -qr $$dir.zip $$dir); fi; \
	done
	cd dist && sha256sum *.zip *.tar.gz > SHA256SUMS.txt
	@ls -l dist/*.zip dist/*.tar.gz

clean:
	rm -rf golog dist web/dist
