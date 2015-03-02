APP_NAME = ceph-metrics
VERSION = latest
BINTRAY_USER = darkcrux
BINTRAY_REPO = generic


all: clean build

clean:
	@echo "--> Cleaning build"
	@go clean -i
	@rm -rf ./build

prepare:
	@mkdir -p build/bin/linux-amd64
	@mkdir -p build/test
	@mkdir -p build/doc
	@mkdir -p build/tar

deps:
	@echo "--> Getting dependecies"
	@go get -v ./...

format:
	@echo "--> Formatting source code"
	@go fmt ./...

test: prepare deps format
	@echo "--> Testing application"
	@go test -v -outputdir build/test ./...

build: test
	@echo "--> Building local application"
	@go build -o build/bin/linux-amd64/${VERSION}/${APP_NAME} -v .

install: build
	@echo "--> Installing application"
	@go install

package: build
	@echo "--> Packaging application"
	@tar cfz build/tar/${APP_NAME}-${VERSION}-linux-amd64.tar.gz -C build/bin/linux-amd64/${VERSION} ${APP_NAME}

release: package
ifeq ($(VERSION) , latest)
	@echo "--> Removing Latest Version"
	@curl -s -X DELETE -u ${ACCESS_KEY} https://api.bintray.com/packages/${BINTRAY_USER}/${BINTRAY_REPO}/${APP_NAME}/versions/${VERSION}
	@echo
endif
	@echo "--> Releasing version: ${VERSION}"
	@curl -s -T "build/tar/${APP_NAME}-${VERSION}-linux-amd64.tar.gz" -u "${ACCESS_KEY}" "https://api.bintray.com/content/${BINTRAY_USER}/${BINTRAY_REPO}/${APP_NAME}/${VERSION}/${APP_NAME}-${VERSION}-linux-amd64.tar.gz"
	@echo "... linux-amd64"
	@echo "--> Publishing version ${VERSION}"
	@curl -s -X POST -u ${ACCESS_KEY} https://api.bintray.com/content/${BINTRAY_USER}/${BINTRAY_REPO}/${APP_NAME}/${VERSION}/publish
	@echo 
