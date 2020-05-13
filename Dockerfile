FROM golang:1.14
WORKDIR /go/src/gitlab.com/simonkrenger/avelon-cloud-prometheus-exporter
COPY * ./
COPY avelon/ avelon/
RUN go mod download
# http://blog.wrouesnel.com/articles/Totally%20static%20Go%20builds/
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' main.go

FROM scratch
LABEL maintainer=simon@krenger.ch
WORKDIR /
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /go/src/gitlab.com/simonkrenger/avelon-cloud-prometheus-exporter/main ./avelon-cloud-prometheus-exporter

EXPOSE 8080
USER 1001
CMD ["./avelon-cloud-prometheus-exporter"]
