FROM golang:alpine AS builder

RUN apk add --no-cache curl
RUN apk add --no-cache git
RUN mkdir -p $$GOPATH/bin && \
    curl https://glide.sh/get | sh

ADD . /go/src/minio-cleaner
WORKDIR /go/src/minio-cleaner
RUN glide update && glide install
RUN go build -o minio-cleaner .
RUN chmod +x minio-cleaner

#create new clean image
FROM alpine
WORKDIR /app
COPY --from=builder /go/src/minio-cleaner/minio-cleaner /app/minio-cleaner
RUN cp minio-cleaner /usr/bin/minio-cleaner