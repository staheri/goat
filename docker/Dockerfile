FROM ubuntu:latest
ENV TZ=Europe/Athens
ARG DEBIAN_FRONTEND=noninteractive

RUN apt update
RUN apt install -y software-properties-common
RUN apt install -y\
  wget\
  build-essential\
  manpages-dev\
  graphviz\
  vim\
  git\
  bash\
  sudo\
  openssh-server

SHELL ["/bin/bash", "-c"] 


## Setup SSH connections.
RUN useradd -rm -d /home/user -s /bin/bash -g root -G sudo -u 1000 user
RUN echo 'user:user' | chpasswd
RUN service ssh start

EXPOSE 22

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
RUN git clone https://github.com/goodmorning-coder/gobench.git
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
RUN git remote set-url origin https://github.com/fivosts/goat.git
RUN git pull
RUN export GOPATH=$HOME/gopath;\
  export GOROOT=/usr/local/go;\
  export PATH=$PATH:/usr/local/go/bin:$HOME/gopath/bin;\
  export GOATWS=$HOME/goatws;\
  export GOATTO=30;\
  export GOATMAXPROCS=4;\
  go build -o /root/gopath/bin/goat

# Setup python
RUN add-apt-repository ppa:deadsnakes/ppa
RUN apt update
RUN apt install -y \
    ca-certificates \
    curl \
    python3.8 \
    python3.8-dev \
    python3-distutils \
    python3-numpy
RUN curl https://bootstrap.pypa.io/get-pip.py -o /tmp/get-pip.py
RUN python3 /tmp/get-pip.py
RUN ln -sf /usr/bin/python3 /usr/bin/python

CMD ["/bin/bash", "-c", "bash", ";", "source", "/root/.bashrc", ";", "/usr/sbin/sshd","-D"]
