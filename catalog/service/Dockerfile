FROM obraun/golang-microservices-ci as builder
COPY . /app
WORKDIR /app
RUN go build -o catalog/service/catalogservice catalog/service/main.go

FROM alpine
COPY --from=builder /app/catalog/service/catalogservice /app/catalogservice
EXPOSE 8091
ENTRYPOINT [ "/app/catalogservice" ]