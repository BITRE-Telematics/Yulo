FROM golang:latest
LABEL maintainer="Telematics@infrastructure.gov.au"
RUN mkdir /queries

COPY ./ /queries
RUN chmod 777 -R /queries



WORKDIR /queries
RUN go mod init github.com/bitre-telematics/queries
RUN go mod tidy
RUN go mod download

ENTRYPOINT [ "go", "run", "seg_query.go"]
