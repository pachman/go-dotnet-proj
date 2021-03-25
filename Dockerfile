FROM golang:alpine as builder

WORKDIR /build
COPY . .

RUN go build

FROM alpine/git as runtime

WORKDIR /app

COPY --from=builder /build/go-dotnet-proj .

ENTRYPOINT ./go-dotnet-proj
