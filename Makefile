all:
	CGO_ENABLED=0 go build
clean:
	rm -f mec-ng
