FROM golang:1.17.6-alpine AS build

WORKDIR /app

COPY . /app/

RUN go build -o /deviation_rest

FROM alpine:3.14 AS deploy

COPY --from=build /deviation_rest /deviation_rest

ENTRYPOINT [ "/deviation_rest" ]
