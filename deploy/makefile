
J2GO_FLAG := --outdir=../autogen --dispatch-reporter=true
GOOS      := linux

GO_MOD:=github.com/tarscloud/gopractice
J2GO_FLAG += --module=${GO_MOD}/apps/autogen

include ../../deploy/makefile.tars

IMG=$(shell echo ${TARGET} | tr A-Z a-z)

img: export GOOS=linux
img: $(TARGET)
	docker build -f ../../deploy/Dockerfile -t ${IMG} --build-arg SERVER=${TARGET} .
