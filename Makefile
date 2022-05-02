
all:
	echo "all"

imports:
	@if  ! command openshift-goimports; then \
		echo "installing openshift-goimports"; go install github.com/openshift-eng/openshift-goimports@latest; \
	fi
	openshift-goimports -m github.com/hongkailiu/hello-go
.PHONY: imports


.PHONY:
generate: imports

verify: generate
	@# Don't add --quiet here, it disables --exit code in the git 1.7 we have in CI, making this unusuable
	if  ! git diff --exit-code; then \
    	echo "generated files are out of date, run make generate"; exit 1; \
    fi
.PHONY: verify
