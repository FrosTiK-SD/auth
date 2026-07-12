FROM golang:1.22

ARG ATLAS_URI
ARG FIREBASE_PROJECT_ID
ARG REDIS_URI
ARG REDIS_PASSWORD
ARG REDIS_USERNAME

ENV APP_HOME /go/src/authv2
ENV ATLAS_URI ${ATLAS_URI}
ENV REDIS_URI ${REDIS_URI}
ENV REDIS_PASSWORD ${REDIS_PASSWORD}
ENV REDIS_USERNAME ${REDIS_USERNAME}
ENV FIREBASE_PROJECT_ID ${FIREBASE_PROJECT_ID}
ENV GIN_MODE=release

WORKDIR "$APP_HOME"

RUN go env -w GOPRIVATE=github.com/FrosTiK-SD/*

COPY go.mod .
COPY go.sum .

RUN --mount=type=secret,id=gh_token \
    git config --global url."https://x-access-token:$(cat /run/secrets/gh_token)@github.com/".insteadOf "https://github.com/" \
    && go mod download

COPY . .
RUN go build -tags=jsoniter -o authv2 

EXPOSE 8080

CMD ["./authv2"]
