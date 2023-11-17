# builds service executable
.PHONY: build
build:
	go build -v -o ./bin/goifs pkg/main.go

clean:
	rm -rvf bin build

update:
	go get -u github.com/homebackend/go-homebackend-common

debian: build
	mkdir -p build/debian/goifs/DEBIAN \
		build/debian/goifs/usr/local/sbin \
		build/debian/goifs/etc/goifs/ \
		build/debian/goifs/lib/systemd/system
	cp -v config/ifs.yaml.sample build/debian/goifs/etc/goifs
	cp -v bin/goifs build/debian/goifs/usr/local/sbin
	cp -v goifs.service build/debian/goifs/lib/systemd/system
	cp -v debian.control build/debian/goifs/DEBIAN/control
	dpkg-deb --build build/debian/goifs

debian-install: debian
	sudo dpkg -i build/debian/goifs.deb

# runs the service locally using the credentials provided by aws-vault for dev-00
.PHONY: run
run: build
	@./bin/goifs -c config/ifs.yaml
