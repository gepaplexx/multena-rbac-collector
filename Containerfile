FROM golang:1.20-alpine as BUILD

WORKDIR /app/
COPY go.mod go.sum ./
RUN go mod verify
COPY . .
RUN go build -o rbac-collector ./cmd
RUN  chmod -R g=u /app/rbac-collector

FROM gcr.io/distroless/static

COPY --from=BUILD /app/config/config.yaml /config/config.yaml
COPY --from=BUILD /app/rbac-collector /rbac-collector

ENTRYPOINT ["/rbac-collector"]

