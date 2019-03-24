# Copyright 2017 the Heptio Ark contributors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

BINS = $(wildcard velero-*)

REPO ?= github.com/fusor/ocp-velero-plugin

BUILD_IMAGE ?= openshift/origin-release:golang-1.11

IMAGE ?= docker.io/fusor/ocp-velero-plugin

ARCH ?= amd64

BUILDTAGS ?= containers_image_ostree_stub exclude_graphdriver_devicemapper exclude_graphdriver_btrfs containers_image_openpgp exclude_graphdriver_overlay

all: $(addprefix build-, $(BINS))

build-%:
	$(MAKE) --no-print-directory BIN=$* build

build: _output/$(BIN)

_output/$(BIN): $(BIN)/*.go
	mkdir -p .go/src/$(REPO) .go/pkg .go/std/$(ARCH) _output
	cp -rp * .go/src/$(REPO)
# The patch and mv below is needed until https://github.com/containers/storage/pull/309 is merged and vendor/ is updated with the fix
	patch  -p0 -d .go/src/$(REPO) < copy_cgo.patch
	mv .go/src/$(REPO)/vendor/github.com/containers/storage/drivers/copy/copy.go .go/src/$(REPO)/vendor/github.com/containers/storage/drivers/copy/copy_linux.go
	docker run \
				 --rm \
				 -u $$(id -u):$$(id -g) \
				 -v $$(pwd)/.go/pkg:/go/pkg \
				 -v $$(pwd)/.go/src:/go/src \
				 -v $$(pwd)/.go/std:/go/std \
				 -v $$(pwd)/_output:/go/src/$(REPO)/_output \
				 -v $$(pwd)/.go/std/$(ARCH):/usr/local/go/pkg/linux_$(ARCH)_static \
				 -e CGO_ENABLED=0 \
				 -w /go/src/$(REPO) \
				 $(BUILD_IMAGE) \
				 go build -installsuffix "static"  -tags "$(BUILDTAGS)" -i -v -o _output/$(BIN) ./$(BIN)

container: all
	cp Dockerfile _output/Dockerfile
	docker build -t $(IMAGE) -f _output/Dockerfile _output

all-ci: $(addprefix ci-, $(BINS))

ci-%:
	$(MAKE) --no-print-directory BIN=$* ci

ci:
	mkdir -p _output
	CGO_ENABLED=0 go build -v -o _output/$(BIN) ./$(BIN)

clean:
	rm -rf .go _output
