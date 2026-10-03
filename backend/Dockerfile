FROM golang:1.26-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/tutorina-api ./cmd/api
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/tutorina-migrate ./cmd/migrate
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/tutorina-key ./cmd/key

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=build /out/tutorina-api ./tutorina-api
COPY --from=build /out/tutorina-migrate ./tutorina-migrate
COPY --from=build /out/tutorina-key ./tutorina-key
COPY --from=build /app/migrations ./migrations

RUN addgroup -S tutorina && adduser -S -G tutorina tutorina
USER tutorina

EXPOSE 8081
CMD ["./tutorina-api"]
