
.PHONY: install
install:
	go install \
	github.com/timtadh/sfp \
	github.com/timtadh/sfp/afp \
	github.com/timtadh/sfp/cmd/clean-go-pprof \
	github.com/timtadh/sfp/cmd/dot-to-veg \
	github.com/timtadh/sfp/cmd/find-embeddings \
	github.com/timtadh/sfp/cmd/list-embeddings
