FROM golang:1.22-alpine AS build
WORKDIR /src
COPY agent /src
RUN go build -o /out/agent ./main.go

FROM alpine:3.20
COPY --from=build /out/agent /usr/local/bin/agent
ENTRYPOINT ["/usr/local/bin/agent"]
