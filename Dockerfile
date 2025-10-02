FROM golang:1.24.7-alpine3.22 AS build



WORKDIR /app

COPY . .

RUN go build -o main .



FROM jrottenberg/ffmpeg:7.1-nvidia2204 AS ffmpeg 






FROM ubuntu:24.04



WORKDIR /app



#  Copy only what you need from the ffmpeg image


COPY --from=ffmpeg   / /


RUN apt-get update && apt-get install -y \
    ca-certificates \
    libnvidia-encode-550 \
    wget \
    && rm -rf /var/lib/apt/lists/*

#  Download and install shaka-packager
RUN wget https://github.com/shaka-project/shaka-packager/releases/download/v3.4.2/packager-linux-x64 && \
    chmod +x packager-linux-x64 && \
    mv packager-linux-x64 /usr/local/bin/packager

#  Copy your Go binary

ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/lib


ENV NVIDIA_VISIBLE_DEVICES=all 
ENV NVIDIA_DRIVER_CAPABILITIES=compute,utility,video 

COPY --from=build /app/main  /app/


CMD ["./main"]