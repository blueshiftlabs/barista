#!/usr/bin/env bash

# From https://github.com/codecov/example-go#caveat-multiple-files
set -e
echo "" > coverage.txt

# Run tests with coverage for all barista packages
for d in $(go list ./... | grep -v barista/samples); do
	go test -coverprofile=profile.out -race -covermode=atomic $d
	if [ -f profile.out ]; then
		cat profile.out >> coverage.txt
		rm profile.out
	fi
done

# Debug log tests need the build tag, otherwise the nop versions will be used.
go test -tags debuglog -coverprofile=profile.out -race -covermode=atomic ./logging
if [ -f profile.out ]; then
	cat profile.out >> coverage.txt
	rm profile.out
fi

# Run tests only for samples.
# This is just to make sure that all samples compile.
go test ./samples/...
