# Compilación
FROM golang:1.25-alpine AS build

WORKDIR /src

# Las dependencias se copian primero para que Docker cachee esta capa: el
# go.mod cambia mucho menos que el código.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 produce un binario estático, requisito para correr sobre
# distroless. -w -s saca la información de debug y baja el tamaño.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/api ./cmd/api

# Imagen final
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /bin/api /api

# nonroot ya viene definido en la imagen base. Un contenedor que corre como
# root es una escalada de privilegios servida.
USER nonroot:nonroot

EXPOSE 8080
ENV APP_ENV=production

ENTRYPOINT ["/api"]
