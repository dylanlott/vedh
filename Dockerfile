# Build image
FROM golang:1.17-bullseye AS build
WORKDIR /app
COPY ./ ./
RUN make build
CMD ["./edhgo"]

# Prod image
FROM golang:1.17-alpine
WORKDIR /app
COPY --from=build /app/edhgo .
EXPOSE 8080
CMD ["./edhgo"]
