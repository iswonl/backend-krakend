FROM golang:1.19-alpine as builder
RUN apk add make gcc musl-dev
WORKDIR /home/golang/plugins/

COPY ./plugins/ ./
RUN cd google-auth && go build -buildmode=plugin -o ../bin/google-auth.so .
RUN cd google-code && go build -buildmode=plugin -o ../bin/google-code.so .
RUN cd auth-token && go build -buildmode=plugin -o ../bin/auth-token.so .

FROM devopsfaith/krakend:2.1.2
COPY ./config/krakend.json /etc/krakend/krakend.json
COPY --from=builder /home/golang/plugins/bin/ /etc/krakend/plugins/
EXPOSE 8080