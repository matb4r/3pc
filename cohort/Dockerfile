FROM golang:1.9.2-alpine3.6
RUN apk update
RUN apk add git
RUN go get github.com/matb4r/3pc/commons
RUN go get github.com/matb4r/3pc/cohort
WORKDIR /go/src/github.com/matb4r/3pc/cohort
ENTRYPOINT ["go","run","cohort.go"]
