# based on http://carlosbecker.com/posts/small-go-apps-containers/
FROM alpine:3.2

ENV GOROOT=/usr/lib/go \
    GOPATH=/gopath \
    GOBIN=/gopath/bin \
    PATH=$PATH:$GOROOT/bin:$GOPATH/bin \
    AUTH_SERVE_ADDRESS=:80 \
    AUTH_SESSION_KEY=auth

WORKDIR /gopath/src/github.com/hashtock/auth
ADD . /gopath/src/github.com/hashtock/auth

RUN apk add -U git go && \
    go get github.com/tools/godep && \
    $GOBIN/godep go build -o /usr/bin/auth && \
    apk del git go && \
    rm -rf /gopath && \
    rm -rf /var/cache/apk/*

EXPOSE 80

CMD "auth"
