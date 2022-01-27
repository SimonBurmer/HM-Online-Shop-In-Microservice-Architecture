FROM obraun/golang-microservices-ci as builder
COPY . /app
WORKDIR /app
RUN go build -o client/service/client client/service/main.go

FROM alpine
COPY --from=builder /app/client/service/client /app/client
ENTRYPOINT [ "/app/client" ]

# Use --network="host" in your docker run command, then 127.0.0.1 in your docker container will point to your docker host.

# docker build -t gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/client  -f client/service/Dockerfile .
# docker build -t client  -f client/service/Dockerfile .