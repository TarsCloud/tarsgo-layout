GO_MOD:=github.com/tarscloud/gopractice
TARS_SRC  := $(wildcard proto/*.tars)
J2GO_FLAG := --outdir=proto/stub --dispatch-reporter=true --module=${GO_MOD}/apps/autogen
GOOS      := linux

include ../../deploy/makefile.tars

IMG=$(shell echo ${TARGET} | tr A-Z a-z)

img: export GOOS=linux
img: $(TARGET)
	docker build -f ../../deploy/Dockerfile -t ${IMG} --build-arg SERVER=${TARGET} .
