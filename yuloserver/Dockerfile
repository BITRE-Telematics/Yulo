FROM golang:latest
LABEL maintainer="Telematics@infrastructure.gov.au"
RUN mkdir /yuloserver

COPY ./ /yuloserver
RUN chmod 777 -R /yuloserver


RUN apt-get update && \
apt-get install -y libgeos-dev && \
apt-get install -y netcat
WORKDIR /yuloserver
RUN go mod init github.com/bitre-telematics/yulo
RUN go mod tidy
RUN go mod download

CMD [ "go", "run", "main.go" ]
