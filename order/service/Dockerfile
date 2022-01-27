FROM obraun/golang-microservices-ci as builder
COPY . /app
WORKDIR /app
RUN go build -o order/service/orderservice order/service/main.go

FROM alpine
COPY --from=builder /app/order/service/orderservice /app/orderservice
ENTRYPOINT [ "/app/orderservice" ]
