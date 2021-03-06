VERSION= 0.7.0
PACKAGES= $(shell find . -name '*.go' -print0 | xargs -0 -n1 dirname | sort --unique)
LDFLAGS= -ldflags "-X main.version=${VERSION}"
DEBPATH= scripts/dpkg
RPMPATH= scripts/rpmbuild
ARCH=`uname -m`
BUILD_OS := $(shell uname -s)

default: test

test:
ifeq ($(BUILD_OS),Darwin)
	go test -cover -count=1 `go list ./... | grep -v "vflow/vflow" | grep -v "mirror"` -timeout 1m
endif
ifeq ($(BUILD_OS),Linux)
	go test -cover -count=1 `go list ./... | grep -v "mirror"` -timeout 1m
endif

bench:
	go test -v ./... -bench=. -timeout 2m

run: build
	cd vflow; ./vflow -sflow-workers 100 -ipfix-workers 100

debug: build
	cd vflow; ./vflow -sflow-workers 100 -ipfix-workers 100 -verbose=true

gctrace: build
	cd vflow; env GODEBUG=gctrace=1 ./vflow -sflow-workers 100 -ipfix-workers 100

lint:
	golint ./...

cyclo:
	gocyclo -over 15 $(PACKAGES)

errcheck:
	errcheck ./...

tools:
	go get github.com/golang/lint/golint
	go get github.com/kisielk/errcheck
	go get github.com/alecthomas/gocyclo

depends:
	go get -d ./...

build: depends
	cd vflow; go build $(LDFLAGS)
	cd stress; go build

dpkg: build
	mkdir -p ${DEBPATH}/etc/init.d ${DEBPATH}/etc/logrotate.d
	mkdir -p ${DEBPATH}/etc/vflow ${DEBPATH}/usr/share/doc/vflow
	mkdir -p ${DEBPATH}/usr/bin ${DEBPATH}/usr/local/vflow
	sed -i 's/%VERSION%/${VERSION}/' ${DEBPATH}/DEBIAN/control
	cp vflow/vflow ${DEBPATH}/usr/bin/
	cp stress/stress ${DEBPATH}/usr/bin/vflow_stress
	cp scripts/vflow.service ${DEBPATH}/etc/init.d/vflow
	cp scripts/vflow.logrotate ${DEBPATH}/etc/logrotate.d/vflow
	cp scripts/vflow.conf ${DEBPATH}/etc/vflow/vflow.conf
	cp scripts/kafka.conf ${DEBPATH}/etc/vflow/mq.conf
	cp scripts/ipfix.elements ${DEBPATH}/etc/vflow/
	cp ${DEBPATH}/DEBIAN/copyright ${DEBPATH}/usr/share/doc/vflow/
	cp LICENSE ${DEBPATH}/usr/share/doc/vflow/license
	dpkg-deb -b ${DEBPATH}
	mv ${DEBPATH}.deb scripts/vflow-${VERSION}-${ARCH}.deb
	sed -i 's/${VERSION}/%VERSION%/' ${DEBPATH}/DEBIAN/control

rpm: build
	sed -i 's/%VERSION%/${VERSION}/' ${RPMPATH}/SPECS/vflow.spec
	rm -rf ${RPMPATH}/SOURCES/
	mkdir ${RPMPATH}/SOURCES/
	cp vflow/vflow ${RPMPATH}/SOURCES/
	cp stress/stress ${RPMPATH}/SOURCES/vflow_stress
	cp scripts/vflow.conf ${RPMPATH}/SOURCES/
	cp scripts/vflow.service ${RPMPATH}/SOURCES/
	cp scripts/vflow.logrotate ${RPMPATH}/SOURCES/
	cp scripts/kafka.conf ${RPMPATH}/SOURCES/mq.conf
	cp scripts/ipfix.elements ${RPMPATH}/SOURCES/
	cp LICENSE ${RPMPATH}/SOURCES/license
	cp NOTICE ${RPMPATH}/SOURCES/notice
	apt-get install rpm
	rpmbuild -ba ${RPMPATH}/SPECS/vflow.spec --define "_topdir `pwd`/scripts/rpmbuild"
	sed -i 's/${VERSION}/%VERSION%/' ${RPMPATH}/SPECS/vflow.spec

###########################################################################
# sonar-scanner configuration section
ifdef ghprbSourceBranch
	SONAR_BRANCH := -Dsonar.pullrequest.branch=$(ghprbSourceBranch) -Dsonar.pullrequest.key=$(ghprbPullId) -Dsonar.pullrequest.base=$(ghprbTargetBranch)
else
	SONAR_BRANCH := -Dsonar.branch.name=$(shell git rev-parse --abbrev-ref HEAD)
endif

SONAR_EXE=sonar-scanner
SONAR_PRO=-Dproject.settings=sonar-project.properties
SONAR_BAS=-Dsonar.projectBaseDir=.
SONAR_CMD=${SONAR_EXE} ${SONAR_PRO} ${SONAR_BAS} ${SONAR_BRANCH}
sonar:
	rm -rf $(BUILD_DIR)
	rm -rf $(LOCAL_GOPATH_DIR)
	${SONAR_CMD}
###########################################################################