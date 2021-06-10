FROM registry.fedoraproject.org/fedora-minimal:34 as build
WORKDIR /go/src/gitlab.com/simonkrenger/avelon-cloud-prometheus-exporter
RUN microdnf install -y golang && go get github.com/prometheus/client_golang
COPY * ./
COPY avelon/ avelon/
RUN go mod download
# http://blog.wrouesnel.com/articles/Totally%20static%20Go%20builds/
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' main.go

FROM scratch
LABEL maintainer="Simon Krenger <simon@krenger.ch>"
WORKDIR /
COPY --from=0 /etc/ssl/certs/ca-bundle.crt /etc/ssl/certs/
COPY --from=0 /go/src/gitlab.com/simonkrenger/avelon-cloud-prometheus-exporter/main ./avelon-cloud-prometheus-exporter

EXPOSE 8080
USER 1001
CMD ["./avelon-cloud-prometheus-exporter"]
