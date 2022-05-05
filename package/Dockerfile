FROM ubuntu:20.04

ARG ARCH=amd64

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get dist-upgrade -y && \
    apt-get install -y kmod curl nfs-common fuse \
    libibverbs1 librdmacm1 libconfig-general-perl libaio1 sg3-utils \
    iputils-ping telnet iperf qemu-utils wget iproute2 && \

COPY ./nvme-go /usr/local/bin

VOLUME /usr/local/bin
