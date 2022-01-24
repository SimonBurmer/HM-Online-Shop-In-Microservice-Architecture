FROM obraun/golang-microservices-ci as builder
COPY . /app
WORKDIR /app
RUN go build -o shipment/service/shipmentservice shipment/service/main.go

FROM alpine
COPY --from=builder /app/shipment/service/shipmentservice /app/shipmentservice
ENTRYPOINT [ "/app/shipmentservice" ]