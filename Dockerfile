# Build image
FROM golang:1.22-bullseye AS build
WORKDIR /app
COPY ./ ./
RUN make build-linux
CMD ["./edhgo_unix"]

# Prod image
FROM golang:1.22-alpine
WORKDIR /app
COPY --from=build /app/edhgo_unix edhgo_unix
EXPOSE 8080
CMD ["./edhgo_unix"]
