FROM golang:latest AS compiling
RUN mkdir -p /go/src/pipeline
WORKDIR /go/src/pipeline
ADD main.go .
ADD go.mod .
#RUN go install .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
#RUN apk --no-cache add ca-certificates
LABEL version="1.1"
LABEL maintainer="Nikolai Uvarov<***@gmail.com>"
WORKDIR /root/
COPY --from=compiling /go/src/pipeline/app .
#RUN ls
#ENTRYPOINT ./website
#CMD ["./app"]
#CMD ["/bin/sh"]
CMD ["./app"]
#EXPOSE 8080