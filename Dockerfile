FROM golang:1.20-alpine

# Instala as dependências necessárias para o SQLite e habilita o CGO
RUN apk add --no-cache gcc musl-dev

# Habilita o CGO para usar o go-sqlite3
ENV CGO_ENABLED=1

WORKDIR /app

COPY . .

RUN go mod download

RUN apk add --no-cache gcc musl-dev curl

EXPOSE 8080

CMD ["go", "run", "server/server.go"]
