FROM golang:1.20

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app .

ENV ADDR="localhost" PORT=3306 PASSWD="YOUR_DB_PASSWORD" MASTER="master" JWT="YOUR_PRIVATE_JWT_KEY"


CMD ["/bin/bash","-c","app -a ${ADDR} -port ${PORT} -pwd ${PASSWD} -Rname ${MASTER} -jwtkey ${JWT}"]