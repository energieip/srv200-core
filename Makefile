COMPONENT=energieip-srv200-core

BINARIES=bin/$(COMPONENT)-armhf bin/$(COMPONENT)-amd64

.PHONY: $(BINARIES) clean

bin/$(COMPONENT)-amd64:
	go build -o $@

bin/$(COMPONENT)-armhf:
	env GOOS=linux GOARCH=arm go build -o $@

all: $(BINARIES)

prepare:
	glide install

deb-amd64: bin/$(COMPONENT)-amd64
	$(eval VERSION := $(shell cat ./version))
	$(eval ARCH := $(shell echo "amd64"))
	$(eval BUILD_NAME := $(shell echo "$(COMPONENT)-$(VERSION)-$(ARCH)"))
	$(eval BUILD_PATH := $(shell echo "build/$(BUILD_NAME)"))
	make deb VERSION=$(VERSION) BUILD_PATH=$(BUILD_PATH) ARCH=$(ARCH) BUILD_NAME=$(BUILD_NAME)


.ONESHELL:
deb-armhf: bin/$(COMPONENT)-armhf
	$(eval VERSION := $(shell cat ./version))
	$(eval ARCH := $(shell echo "armhf"))
	$(eval BUILD_NAME := $(shell echo "$(COMPONENT)-$(VERSION)-$(ARCH)"))
	$(eval BUILD_PATH := $(shell echo "build/$(BUILD_NAME)"))
	make deb VERSION=$(VERSION) BUILD_PATH=$(BUILD_PATH) ARCH=$(ARCH) BUILD_NAME=$(BUILD_NAME)

deb:
	mkdir -p $(BUILD_PATH)/usr/local/bin $(BUILD_PATH)/etc/$(COMPONENT) $(BUILD_PATH)/etc/systemd/system $(BUILD_PATH)/data/www/ $(BUILD_PATH)/usr/local/share/fonts/
	cp -r ./scripts/DEBIAN $(BUILD_PATH)/
	cp ./scripts/config.json $(BUILD_PATH)/etc/$(COMPONENT)/
	cp ./scripts/*.service $(BUILD_PATH)/etc/systemd/system/
	sed -i "s/amd64/$(ARCH)/g" $(BUILD_PATH)/DEBIAN/control
	sed -i "s/VERSION/$(VERSION)/g" $(BUILD_PATH)/DEBIAN/control
	sed -i "s/COMPONENT/$(COMPONENT)/g" $(BUILD_PATH)/DEBIAN/control
	cp ./scripts/Makefile $(BUILD_PATH)/../
	cp bin/$(COMPONENT)-$(ARCH) $(BUILD_PATH)/usr/local/bin/$(COMPONENT)
	cp cmd/*.py $(BUILD_PATH)/usr/local/bin/
	cp cmd/*.ttf $(BUILD_PATH)/usr/local/share/fonts/
	mkdir -p $(BUILD_PATH)/usr/local/share/images/
	cp cmd/*.png $(BUILD_PATH)/usr/local/share/images/
	chmod +x $(BUILD_PATH)/usr/local/bin/*.py
	cp -r internal/swaggerui $(BUILD_PATH)/data/www/
	cp -r internal/swagger/*.json $(BUILD_PATH)/data/www/swaggerui/
	make -C build DEB_PACKAGE=$(BUILD_NAME) deb

clean:
	rm -rf bin build

run: bin/$(COMPONENT)-amd64
	./bin/$(COMPONENT)-amd64 -c ./scripts/config.json
