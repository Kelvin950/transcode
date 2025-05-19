FROM  golang:1.24.3-alpine3.20  AS build

WORKDIR  /app 

COPY . . 

RUN  go build -o main .




FROM alpine

WORKDIR /app 


 RUN apk add ffmpeg 


COPY --from=build  /app/main  . 


CMD [ "./main" ]