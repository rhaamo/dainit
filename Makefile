.PHONY: build

build: build-lutrainit build-lutractl
build-lutrainit:
	$(MAKE) -C lutrainit build

build-lutractl:
	$(MAKE) -C lutractl build

build-race: build-race-lutrainit build-race-lutractl
build-race-lutrainit:
	$(MAKE) -C lutrainit build-race

build-race-lutractl:
	$(MAKE) -C lutractl build-race

vet: vet-lutrainit vet-lutractl
vet-lutrainit:
	$(MAKE) -C lutrainit vet

vet-lutractl:
	$(MAKE) -C lutractl vet

clean: clean-lutrainit clean-lutractl
clean-lutrainit:
	$(MAKE) -C lutrainit clean

clean-lutractl:
	$(MAKE) -C lutractl clean

test: test-lutrainit test-lutractl
test-lutrainit:
	$(MAKE) -C lutrainit test

test-lutractl:
	$(MAKE) -C lutractl test

misspell-check: misspell-check-lutrainit misspell-check-lutractl
misspell-check-lutrainit:
	$(MAKE) -C lutrainit misspell-check

misspell-check-lutractl:
	$(MAKE) -C lutractl misspell-check

fmt-check: fmt-check-lutrainit fmt-check-lutractl
fmt-check-lutrainit:
	$(MAKE) -C lutrainit fmt-check

fmt-check-lutractl:
	$(MAKE) -C lutractl fmt-check

fmt: fmt-lutrainit fmt-lutractl
fmt-lutrainit:
	$(MAKE) -C lutrainit fmt

fmt-lutractl:
	$(MAKE) -C lutractl fmt

lint: lint-lutrainit lint-lutractl
lint-lutrainit:
	$(MAKE) -C lutrainit lint

lint-lutractl:
	$(MAKE) -C lutractl lint

## Other targets
install:
	install -m 0755 -p lutrainit/lutrainit /lutrainit
	install -m 0755 -p lutractl/lutractl /usr/bin/lutractl

install-sample-conf:
	install -d -m 0755 /etc/lutrainit
	install -d -m 0755 /etc/lutrainit.d
	install -m 0755 -p conf/lutra.conf /etc/lutrainit/lutra.conf
	install -m 0755 -p conf/lutra.d/loopback.service /etc/lutrainit/lutra.d/
	install -m 0755 -p conf/lutra.d/network.eth0.service /etc/lutrainit/lutra.d/
	install -m 0755 -p conf/lutra.d/udev.service /etc/lutrainit/lutra.d/
	install -m 0755 -p conf/lutra.d/wpa_supplicant.service /etc/lutrainit/lutra.d/

docker-build: build
	docker build -t dashie/lutrainit:latest .

docker-run: docker-build
	docker run --entrypoint /usr/local/bin/lutrainit --name lutrainit dashie/lutrainit:latest 

docker-rm:
	docker rm lutrainit

