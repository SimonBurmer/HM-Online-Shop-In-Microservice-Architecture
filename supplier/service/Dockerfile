FROM obraun/golang-microservices-ci as builder
COPY . /app
WORKDIR /app
RUN go build -o supplier/service/supplierservice supplier/service/main.go

FROM alpine
COPY --from=builder /app/supplier/service/supplierservice /app/supplierservice
ENTRYPOINT [ "/app/supplierservice" ]
