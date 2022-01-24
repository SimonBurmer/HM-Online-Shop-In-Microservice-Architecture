FROM obraun/golang-microservices-ci as builder
COPY . /app
WORKDIR /app
RUN go build -o payment/service/paymentservice payment/service/main.go

FROM alpine
COPY --from=builder /app/payment/service/paymentservice /app/paymentservice
ENTRYPOINT [ "/app/paymentservice" ]
