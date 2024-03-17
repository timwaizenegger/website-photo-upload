FROM golang:1.22 as builder

WORKDIR /usr/src/app

# Don't need to worry about accumulating data on layers; we'll start a new runtime image later
RUN apt-get update && apt-get install libmagickwand-dev -y

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/website-photo-upload ./...

FROM debian:bookworm-slim

RUN apt-get update && apt-get --no-install-recommends install -y imagemagick && apt-get clean

COPY --from=builder /usr/local/bin/website-photo-upload /usr/local/bin/website-photo-upload
COPY html /app/html
RUN mkdir -p /app/images/thumbs && chown 1100:1100 -R /app
WORKDIR /app

USER 1100
CMD ["website-photo-upload"]


