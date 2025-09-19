BUILDDATE := $(shell date +%Y-%m-%d)
# lazy evaluation of VERSION
VERSION = $(shell cat generated_version)
COMMIT   := $(shell git rev-parse --short=8 HEAD)

LDFLAGS=-ldflags "\
  -s -w \
  -X main.version=${VERSION} \
  -X main.commit=${COMMIT} \
  -X main.buildDate=${BUILDDATE}"

# detect distribution, fallback to bullseye for Jenkins/Docker environments
DISTRIBUTION := $(shell bash -c 'source /etc/os-release 2>/dev/null && echo $$VERSION_CODENAME || echo "bullseye"')

# choose lintian profile based on distro codename
LINTIAN_PROFILE := debian
ifneq (,$(filter $(DISTRIBUTION),bionic focal jammy kinetic lunar mantic noble oracular))
LINTIAN_PROFILE := ubuntu
endif

GO ?= $(firstword $(wildcard /usr/local/go/bin/go) /usr/bin/go)

# pin golangci-lint major version (and exact version) for reproducibility
GOLANGCI_LINT_VERSION ?= v2.3.0

all: build

.PHONY: generate_version
generate_version:
	./generate_version.sh

.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint is not installed. Installing $(GOLANGCI_LINT_VERSION)..." && \
		${GO} install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi
	PATH="$$PATH:$$( ${GO} env GOPATH )/bin" golangci-lint run

.PHONY: build
build: generate_version lint
	${GO} mod tidy
	CGO_ENABLED=0 GOMAXPROCS=1 ${GO} build $(LDFLAGS) -o lelastic

.PHONY: package
package: generate_version lint
	@mkdir -p build
	sed -e "s/##VERSION##/$(shell cat generated_version)/g" \
	-e "s/##DATE##/$(shell TZ=UTC date --date=@0 -R)/g" \
	-e "s/##DISTRIBUTION##/$(DISTRIBUTION)/g" \
	< debian/changelog.template > debian/changelog
	@echo Changelog:
	@cat -n debian/changelog
	dpkg-buildpackage -b -us -uc
	mv ../lelastic*.deb ./
	mv ../lelastic*.buildinfo ./
	mv ../lelastic*.changes ./
	lintian --profile $(LINTIAN_PROFILE) ./*.changes > lintian.log || (cat -n lintian.log && false)
	# @find debian/.debhelper -type f -exec chmod 666 {} \; 2>/dev/null || true
	# @find debian/.debhelper -type d -exec chmod 777 {} \; 2>/dev/null || true
	# @sudo rm -rf debian/.debhelper/ 2>/dev/null || true
	dh_clean

