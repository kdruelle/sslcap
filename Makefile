################################################################################
##
## MIT License
##
## Copyright (c) 2017 Kevin Druelle
##
## Permission is hereby granted, free of charge, to any person obtaining a copy
## of this software and associated documentation files (the "Software"), to deal
## in the Software without restriction, including without limitation the rights
## to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
## copies of the Software, and to permit persons to whom the Software is
## furnished to do so, subject to the following conditions:
##
## The above copyright notice and this permission notice shall be included in all
## copies or substantial portions of the Software.
##
## THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
## IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
## FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
## AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
## LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
## OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
## SOFTWARE
##
###############################################################################


VERSION=$(shell git describe --tags)
BUILDTIME=$(shell LANG=en_US.UTF-8; date +'%d %b %Y')

ARCHS = linux/amd64 linux/386 darwin/amd64 darwin/386 windows/amd64 windows/386
LINUX_ARCHS  = linux/amd64 linux/386
DARWIN_ARCHS = darwin/amd64 darwin/386
WIN_ARCHS    = windows/amd64 windows/386

LINUX_BIN_TARGETS  = $(addprefix bin/sslcap-${VERSION}-,$(subst /,-,${LINUX_ARCHS}))
DARWIN_BIN_TARGETS = $(addprefix bin/sslcap-${VERSION}-,$(subst /,-,${DARWIN_ARCHS}))
WIN_BIN_TARGETS    = $(shell echo $(addprefix bin/sslcap-${VERSION}-,$(subst /,-,${WIN_ARCHS})) | sed -E 's/-windows-([a-z0-9]*)/-windows-\1.exe/g')

LINUX_ZIP_TARGETS  = $(addsuffix .zip, $(LINUX_BIN_TARGETS))
DARWIN_ZIP_TARGETS = $(addsuffix .zip, $(DARWIN_BIN_TARGETS))
WIN_ZIP_TARGETS    = $(shell echo $(addsuffix .zip,$(WIN_BIN_TARGETS)) | sed -E 's/\.exe//g')

BIN_TARGETS = $(LINUX_BIN_TARGETS) $(DARWIN_BIN_TARGETS) $(WIN_BIN_TARGETS)
ZIP_TARGETS = $(LINUX_ZIP_TARGETS) $(DARWIN_ZIP_TARGETS) $(WIN_ZIP_TARGETS)

all:
	go build -ldflags '-X "main.versionStr=${VERSION}" -X "main.buildTime=${BUILDTIME}"'

releases: $(ZIP_TARGETS) $(LINUX_ZIP_TARGETS:%.zip=%.deb)

$(LINUX_ZIP_TARGETS:%.zip=%.deb) : %.deb : %.zip
	@echo "DEB $(notdir $@)"
	export SSLCAP_VERSION=${VERSION}; export ARCH=$(shell echo $< | sed -E 's/.*-([a-z0-9]+).zip$$/\1/g'); \
	sed -i '' 's#.*/usr/local/bin/sslcap.*#    ./$(basename $<): "/usr/local/bin/sslcap"#g' nfpm.yaml ; \
	sed -i '' 's#.*arch.*#arch: "$(shell echo $< | sed -E 's/.*-([a-z0-9]+).zip$$/\1/g')"#g' nfpm.yaml ; \
	nfpm pkg -t $@

$(LINUX_ZIP_TARGETS:%.zip=%.rpm): %.rpm : %.zip
	@echo "RPM $(notdir $@)"
	@export SSLCAP_VERSION=${VERSION}; export ARCH=$(shell echo $< | sed -E 's/.*-([a-z0-9]+).zip$$/\1/g'); \
	sed -i '' 's#.*/usr/local/bin/sslcap.*#    ./$(basename $<): "/usr/local/bin/sslcap"#g' nfpm.yaml ; \
	sed -i '' 's#.*arch.*#arch: "$(shell echo $< | sed -E 's/.*-([a-z0-9]+).zip$$/\1/g')"#g' nfpm.yaml ; \
	nfpm pkg -t $@

$(LINUX_ZIP_TARGETS): %.zip : %
	@echo "Zip $(notdir $@)"
	@cp $< ./bin/sslcap
	@cd ./bin && zip $(notdir $@) sslcap > /dev/null

$(DARWIN_ZIP_TARGETS): %.zip : %
	@echo "Zip $(notdir $@)"
	@cp $< ./bin/sslcap
	@cd ./bin && zip $(notdir $@) sslcap > /dev/null

$(WIN_ZIP_TARGETS): %.zip : %.exe
	@echo "Zip $(notdir $@)"
	@cp $< ./bin/sslcap.exe
	@cd ./bin && zip $(notdir $@) sslcap.exe > /dev/null

$(BIN_TARGETS): $($(wildcard *.go))
	@printf "Build sslcap for target $(subst .exe,,$(subst -,/,$(subst bin/sslcap-${VERSION}-,,$@))) ..."
	@xgo -targets $(subst .exe,,$(subst -,/,$(subst bin/sslcap-${VERSION}-,,$@))) -ldflags '-s -w -X "main.versionStr=${VERSION}" -X "main.buildTime=${BUILDTIME}"' -dest ./bin -out sslcap-${VERSION} ./ > /dev/null
	@find ./bin -name 'sslcap-v*' | sed -e 'p;s/-[0-9.]*-/-/g' | xargs -n2 mv > /dev/null
	@echo "done !"

ci:
	gox -osarch="linux/amd64 linux/386 linux/arm darwin/amd64 darwin/386 windows/amd64 windows/386" -output "./build/sslcap-${VERSION}-{{.OS}}-{{.Arch}}.bin" -ldflags '-s -w -X "main.versionStr=${VERSION}" -X "main.buildTime=${BUILDTIME}"'
	@mkdir -p dist
	@export SSLCAP_VERSION=${VERSION}; ./release.sh
