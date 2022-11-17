FROM ubuntu:latest

RUN apt update
RUN apt install -y\
  wget\
  build-essential\
  manpages-dev\
  graphviz\
  vim\
  git\
  bash

SHELL ["/bin/bash", "-c"] 

## Install everything to /root.
WORKDIR /root/
RUN wget https://golang.org/dl/go1.15.6.linux-amd64.tar.gz
RUN mkdir -p /usr/local/go-goat
RUN tar -xzvf go1.15.6.linux-amd64.tar.gz -C /usr/local
RUN tar -xzvf go1.15.6.linux-amd64.tar.gz -C /usr/local/go-goat
RUN mv /usr/local/go /usr/local/go-orig
RUN ln -nsf /usr/local/go-orig /usr/local/go

## Setup environment variables.
RUN echo "export GOPATH=$HOME/gopath" >> /root/.bashrc
RUN echo "export GOROOT=/usr/local/go" >> /root/.bashrc
RUN echo "export PATH=$PATH:/usr/local/go/bin:$HOME/gopath/bin" >> /root/.bashrc
RUN echo "export GOATWS=$HOME/goatws" >> /root/.bashrc
RUN echo "export GOATTO=30" >> /root/.bashrc
RUN echo "export GOATMAXPROCS=4" >> /root/.bashrc

RUN mkdir /root/gopath
RUN mkdir -p /root/goatws

## Get GOAT tool.
RUN export GOPATH=$HOME/gopath;\
  export GOROOT=/usr/local/go;\
  export PATH=$PATH:/usr/local/go/bin:$HOME/gopath/bin;\
  export GOATWS=$HOME/goatws;\
  export GOATTO=30;\
  export GOATMAXPROCS=4;\
  go get github.com/staheri/goat; exit 0

## Install patch.
WORKDIR /usr/local/go-goat/go
RUN patch -p2 -d src/  < /root/gopath/src/github.com/staheri/goat/go1.15.6_goat_june15.patch; exit 0
WORKDIR /usr/local/go-goat/go/src
RUN export GOPATH=$HOME/gopath;\
  export GOROOT=/usr/local/go;\
  export PATH=$PATH:/usr/local/go/bin:$HOME/gopath/bin;\
  export GOATWS=$HOME/goatws;\
  export GOATTO=30;\
  export GOATMAXPROCS=4;\
  go get github.com/staheri/goat;\
  export GOROOT_BOOTSTRAP=/usr/local/go-orig;\
  ./make.bash
RUN ln -nsf /usr/local/go-goat/go /usr/local/go

WORKDIR /root
RUN rm -rf /root/gopath/src/github.com/rivo/uniseg
RUN wget https://github.com/rivo/uniseg/archive/refs/tags/v0.1.0.tar.gz
RUN tar -xzvf v0.1.0.tar.gz -C /root/gopath/src/github.com/rivo
RUN mv /root/gopath/src/github.com/rivo/uniseg-0.1.0 /root/gopath/src/github.com/rivo/uniseg
RUN export GOPATH=$HOME/gopath;\
  export GOROOT=/usr/local/go;\
  export PATH=$PATH:/usr/local/go/bin:$HOME/gopath/bin;\
  export GOATWS=$HOME/goatws;\
  export GOATTO=30;\
  export GOATMAXPROCS=4;\
  go get github.com/staheri/goat; exit 0

WORKDIR /root/gopath/src/github.com/staheri/goat
RUN grep -rl "/home/saeed" . | xargs sed -i 's/\/home\/saeed/\/root/g'
RUN export GOPATH=$HOME/gopath;\
  export GOROOT=/usr/local/go;\
  export PATH=$PATH:/usr/local/go/bin:$HOME/gopath/bin;\
  export GOATWS=$HOME/goatws;\
  export GOATTO=30;\
  export GOATMAXPROCS=4;\
  go build -o /root/gopath/bin/goat 

CMD ["/bin/bash", "-c", "bash", ";", "source", "/root/.bashrc"]
