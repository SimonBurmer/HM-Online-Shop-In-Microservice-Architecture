FROM obraun/golang-microservices-ci as builder
COPY . /app
WORKDIR /app
RUN go build -o stock/service/stockservice stock/service/main.go

FROM alpine
COPY --from=builder /app/stock/service/stockservice /app/stockservice
EXPOSE 8091
ENTRYPOINT [ "/app/stockservice" ]
