#!/bin/bash

# cd cmd/main && go build -o bdpan && mv bdpan $(go env GOPATH)/bin && cd --
# go build bdpan && mv bdpan $(go env GOPATH)/bin
go build  && mv bdpan-cli $(go env GOPATH)/bin
