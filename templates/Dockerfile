 FROM golang:1.21 AS builder
 WORKDIR /app
 COPY . .
 RUN go mod download
 RUN go build -o {{.ProjectName}} .

 FROM alpine:latest  
 WORKDIR /app/
 COPY --from=builder /app/{{.ProjectName}} .
 
 CMD ["./{{.ProjectName}}"]

