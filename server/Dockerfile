FROM golang:1.26.3-alpine3.23 AS dev

RUN addgroup -S appgroup && adduser -S appuser -G appgroup -h /home/appuser

WORKDIR /app

RUN go install github.com/air-verse/air@v1.65.3

COPY --chown=appuser:appgroup go.mod go.sum ./
RUN go mod download

COPY --chown=appuser:appgroup . .

USER appuser

EXPOSE 8080

CMD ["air"]

# build: compile a static, CGO-free binary. Building on the target host (Arm A1)
# produces a native arm64 binary with no cross-compilation needed.
FROM golang:1.26.3-alpine3.23 AS build

RUN addgroup -S appgroup && adduser -S appuser -G appgroup -h /home/appuser

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /main .

# prod: distroless scratch image holding only the binary + passwd/group so it can
# run as the unprivileged appuser.
FROM scratch AS production

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

WORKDIR /

COPY --from=build --chown=appuser:appgroup /main /main

USER appuser

EXPOSE 8080

ENTRYPOINT ["/main"]
