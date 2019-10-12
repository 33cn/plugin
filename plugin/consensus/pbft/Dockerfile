# This dockerfile uses the ubuntu:18.04 image
# Author: Jhc
# Command format: 

# Base image to use
FROM ubuntu:18.04

# Maintainer
MAINTAINER jhc "jhc@33.cn"

WORKDIR /root

ENV DEBIAN_FRONTEND noninteractive
ENV GOROOT=/usr/local/go
ENV PATH=$PATH:$GOROOT/bin
ENV GOPATH=/root/GO

RUN apt-get update && apt-get install -y apt-utils \
&& apt-get autoclean \
&& apt-get clean

RUN apt-get update && apt-get install -y dialog curl gcc g++ automake autoconf libtool make \
&& curl -o golang.tar.gz "https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz" \
&& echo "1dfe664fa3d8ad714bbd15a36627992effd150ddabd7523931f077b3926d736d golang.tar.gz" | sha256sum -c - \
&& tar -C /usr/local/ -xzf golang.tar.gz \
&& rm -rf golang.tar.gz \
&& mkdir -p /root/GO/src/github.com/33cn/ \
&& apt-get autoclean \
&& apt-get clean

RUN apt-get install -y git \
&& apt-get autoclean \
&& apt-get clean

RUN git clone -b go-pbft \
https://github.com/Jhc-china/plugin.git \
$GOPATH/src/github.com/33cn/plugin

WORKDIR $GOPATH/src/github.com/33cn/plugin/plugin/consensus/pbft
RUN git pull

WORKDIR $GOPATH/src/github.com/33cn/plugin/
RUN make build

WORKDIR $GOPATH/src/github.com/33cn/plugin/build/
ENTRYPOINT [ "./chain33", "-f" ]