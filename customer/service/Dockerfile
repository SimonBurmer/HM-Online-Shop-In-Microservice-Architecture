FROM obraun/golang-microservices-ci as builder
COPY . /app
WORKDIR /app
RUN go build -o customer/service/customerservice customer/service/main.go

FROM alpine
COPY --from=builder /app/customer/service/customerservice /app/customerservice
ENTRYPOINT [ "/app/customerservice" ]
