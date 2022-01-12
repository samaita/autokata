FROM golang:1.14
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go get github.com/gin-gonic/gin
RUN go get github.com/mmcdole/gofeed
RUN go get github.com/robfig/cron
RUN go get github.com/samaita/autokata/sql
RUN go build -o main .
CMD ["/app/main"]