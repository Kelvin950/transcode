FROM  golang:1.24.3-alpine3.20  AS build

WORKDIR  /app 

COPY /snap/bin/ffmpeg  /app/snap/bin/ffmpeg
COPY . . 

RUN  go build -o main .


FROM alpine 

WORKDIR /app 
COPY --from=build  /app/snap/bin/ffmpeg  /app/snap/bin/ffmpeg
COPY --from=build  /app/main  . 


CMD [ "./main" ]