#!/usr/bin/env bash

export GOPATH=$HOME/gopath
mkdir $GOPATH
export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
export GOATWS=$HOME/goatws
mkdir -p $GOATWS
export GOATTO=30
export GOATMAXPROCS=4
