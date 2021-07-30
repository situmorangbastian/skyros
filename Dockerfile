# Start by building the application.
FROM golang:1.16-alpine3.14 as build

WORKDIR /skyros

COPY . .

RUN go mod vendor

RUN CGO_ENABLED=0 go build -o skyros github.com/situmorangbastian/skyros/cmd/skyros

FROM gcr.io/distroless/static

WORKDIR /skyros

COPY --from=build /skyros /app

EXPOSE 4221

CMD ["/app/skyros"]
