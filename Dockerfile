FROM golang:1.24.3-alpine3.20 AS build



WORKDIR /app

COPY . .

RUN go build -o main .



FROM jrottenberg/ffmpeg:7.1-nvidia2204 AS ffmpeg 


FROM google/shaka-packager AS shaka



FROM ubuntu:24.04



WORKDIR /app



# ✅ Copy only what you need from the ffmpeg image


COPY --from=ffmpeg   / /

COPY --from=shaka / /

# ✅ Copy your Go binary


RUN apt-get update && apt-get install -y \
    ca-certificates \
    libnvidia-encode-550 \
    && rm -rf /var/lib/apt/lists/*

ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/lib


ENV NVIDIA_VISIBLE_DEVICES=all 
ENV NVIDIA_DRIVER_CAPABILITIES=compute,utility,video 

COPY --from=build /app/main  /app/


CMD ["./main"]