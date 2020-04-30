FROM quay.io/prometheus/golang-builder AS builder

# Get sql_exporter
<<<<<<< HEAD
ADD .   /go/src/github.com/HBOCodeLabs/sql-exporter
WORKDIR /go/src/github.com/HBOCodeLabs/sql-exporter
=======
ADD .   /go/src/sql_exporter
WORKDIR /go/src/sql_exporter
>>>>>>> 2f68b11ba8fa186d79e8a31c5b4cd32245ad5f99

# Do makefile
RUN make

# Make image and copy build sql_exporter
FROM        quay.io/prometheus/busybox:glibc
MAINTAINER  The Prometheus Authors <prometheus-developers@googlegroups.com>
<<<<<<< HEAD
COPY        --from=builder /go/src/github.com/HBOCodeLabs/sql-exporter/sql_exporter  /bin/sql_exporter
=======
COPY        --from=builder /go/src/sql_exporter/sql_exporter  /bin/sql_exporter
>>>>>>> 2f68b11ba8fa186d79e8a31c5b4cd32245ad5f99

EXPOSE      9399
ENTRYPOINT  [ "/bin/sql_exporter" ]
