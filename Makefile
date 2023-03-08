# Copyright 2023 The Volcano Authors.
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
# 	http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

BIN_DIR=_output/bin
RELEASE_DIR=_output/release
IMAGE_PREFIX=volcanosh
CC ?= "gcc"
BUILDX_OUTPUT_TYPE ?= "docker"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Get OS architecture
OSARCH=$(shell uname -m)
ifeq ($(OSARCH),x86_64)
GOARCH?=amd64
else ifeq ($(OSARCH),x64)
GOARCH?=amd64
else ifeq ($(OSARCH),aarch64)
GOARCH?=arm64
else ifeq ($(OSARCH),aarch64_be)
GOARCH?=arm64
else ifeq ($(OSARCH),armv8b)
GOARCH?=arm64
else ifeq ($(OSARCH),armv8l)
GOARCH?=arm64
else ifeq ($(OSARCH),i386)
GOARCH?=x86
else ifeq ($(OSARCH),i686)
GOARCH?=x86
else ifeq ($(OSARCH),arm)
GOARCH?=arm
else
GOARCH?=$(OSARCH)
endif

# Run `make images DOCKER_PLATFORMS="linux/amd64,linux/arm64" BUILDX_OUTPUT_TYPE=registry IMAGE_PREFIX=[yourregistry]` to push multi-platform
DOCKER_PLATFORMS ?= "linux/${GOARCH}"

GOOS ?= linux

include Makefile.def

.EXPORT_ALL_VARIABLES:

all: vg-controller-manager

init:
	mkdir -p ${BIN_DIR}
	mkdir -p ${RELEASE_DIR}

vg-aggregated-apiserver: init
	CC=${CC} CGO_ENABLED=0 go build -ldflags ${LD_FLAGS} -gcflags="all=-N -l" -o ${BIN_DIR}/vg-aggregated-apiserver ./cmd/aggregated-apiserver

vg-controller-manager: init
	CC=${CC} CGO_ENABLED=0 go build -ldflags ${LD_FLAGS} -gcflags="all=-N -l" -o ${BIN_DIR}/vg-controller-manager ./cmd/controller-manager

vg-scheduler: init
	CC=${CC} CGO_ENABLED=0 go build -ldflags ${LD_FLAGS} -gcflags="all=-N -l" -o ${BIN_DIR}/vg-scheduler ./cmd/global-scheduler

images:
	for name in aggregated-apiserver controller-manager global-scheduler; do\
		docker buildx build -t "${IMAGE_PREFIX}/vg-$$name:$(TAG)" . -f ./installer/dockerfile/$$name/Dockerfile --output=type=${BUILDX_OUTPUT_TYPE} --platform ${DOCKER_PLATFORMS}; \
	done
