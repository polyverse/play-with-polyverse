FROM golang:1.9

COPY . /go/src/github.com/polyverse/play-with-polyverse

WORKDIR /go/src/github.com/polyverse/play-with-polyverse

RUN ssh-keygen -N "" -t rsa -f /etc/ssh/ssh_host_rsa_key >/dev/null

RUN CGO_ENABLED=0 go build -a -installsuffix nocgo -o /go/bin/play-with-polyverse .


FROM alpine

RUN apk --update add ca-certificates
RUN mkdir -p /app/pwd

COPY --from=0 /go/bin/play-with-polyverse /app/play-with-polyverse
COPY --from=0 /etc/ssh/ssh_host_rsa_key /etc/ssh/ssh_host_rsa_key
COPY ./www /app/www

WORKDIR /app
CMD ["./play-with-polyverse"]

EXPOSE 3000
