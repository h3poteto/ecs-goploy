.PHONY: all dep build

OUTPUT = ecs-goploy

all:  build

dep: Gopkg.lock
	dep ensure

build: dep
	go build -o ${OUTPUT} -a -tags netgo -installsuffix netgo --ldflags '-extldflags "-static"'

