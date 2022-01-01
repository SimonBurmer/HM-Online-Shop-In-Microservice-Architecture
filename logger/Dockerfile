FROM obraun/golang-microservices-ci as builder
COPY . /app
WORKDIR /app
RUN go build -o logger/loggerservice logger/logger.go

FROM alpine
COPY --from=builder app/logger/loggerservice /app/loggerservice  
EXPOSE 8091
ENTRYPOINT [ "/app/loggerservice" ]
