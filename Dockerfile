FROM golang:alpine as builder

WORKDIR /build
COPY . .

RUN go build

FROM alpine/git as runtime

ARG SSH_PRIVATE_KEY

WORKDIR /app

RUN mkdir -p ~/.ssh \
    && chmod 0700 ~/.ssh \
    && echo $"$SSH_PRIVATE_KEY">~/.ssh/id_rsa \
    && chmod 0600 ~/.ssh/id_rsa

COPY --from=builder /build/go-dotnet-proj .

ENTRYPOINT ./go-dotnet-proj
