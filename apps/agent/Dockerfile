FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -trimpath -ldflags="-s -w" -o /out/hris-agent .

FROM alpine:3.22
RUN apk --no-cache add ca-certificates && addgroup -S agent && adduser -S -G agent agent
COPY --from=build /out/hris-agent /usr/local/bin/hris-agent
USER agent
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/hris-agent"]
