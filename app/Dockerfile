FROM alpine:3.19

RUN apk --no-cache add gcompat

WORKDIR /usr/src/app

COPY ./cmd/gorilla/gorilla ./

CMD ["/usr/src/app/gorilla"]
