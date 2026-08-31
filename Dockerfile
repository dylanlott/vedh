# Build image
FROM golang:1.27-bookworm AS build
WORKDIR /app
COPY ./ ./
RUN make build-linux
CMD ["./edhgo_unix"]

# Prod image
FROM golang:1.27-alpine
WORKDIR /app
COPY --from=build /app/edhgo_unix edhgo_unix
COPY --from=build /app/persistence/migrations /app/persistence/migrations
EXPOSE 8080
CMD ["./edhgo_unix"]
