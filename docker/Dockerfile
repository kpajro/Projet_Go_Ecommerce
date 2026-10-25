# syntax=docker/dockerfile:1

# ---------- Étape 1 : compilation ----------
FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 produit un binaire statique, exécutable dans une image Alpine
# minimale. -trimpath et -ldflags retirent les chemins de build et réduisent
# la taille du binaire.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/pgo-server ./cmd/server \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/pgo-client ./cmd/client \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/pgo-admin  ./cmd/admin

# ---------- Étape 2 : image finale ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -u 10001 pgo

COPY --from=build /out/ /usr/local/bin/

USER pgo
EXPOSE 8080

ENTRYPOINT ["pgo-server"]
