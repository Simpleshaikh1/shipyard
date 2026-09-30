FROM golang:1.25.0
WORKDIR /src

# Dependencies first: this layer is only rebuilt when go.mod or go.sum change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/api ./cmd/api

EXPOSE 8080
CMD ["api"]