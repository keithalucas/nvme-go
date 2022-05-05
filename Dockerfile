FROM ubuntu:20.04

ARG ARCH=amd64

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get dist-upgrade -y 


COPY nvme-go /usr/local/bin
