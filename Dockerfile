FROM golang:1.16 as build
WORKDIR /go/src/man-counter/
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./counter .


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build /go/src/man-counter/counter ./
CMD ["./counter"]
