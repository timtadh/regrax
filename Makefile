
.PHONY: install
install:
	go install \
	github.com/timtadh/regrax \
	github.com/timtadh/regrax/afp \
	github.com/timtadh/regrax/cmd/clean-go-pprof \
	github.com/timtadh/regrax/cmd/dot-to-veg \
	github.com/timtadh/regrax/cmd/find-embeddings \
	github.com/timtadh/regrax/cmd/list-embeddings
