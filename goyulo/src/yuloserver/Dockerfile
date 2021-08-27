FROM golang:latest 
LABEL maintainer="Richard.Green@infrastructure.gov.au"

ENV GOPATH /./

COPY ./ ./



RUN apt-get update && \
 apt-get install -y libgeos-dev && \
 apt-get install -y netcat

RUN go get github.com/neo4j/neo4j-go-driver/neo4j && \
 go get github.com/paulmach/orb && \
 go get github.com/paulsmith/gogeos/geos && \
 go get github.com/tidwall/gjson && \
 go get github.com/gosexy/to && \
 go get github.com/gosexy/dig && \
 go get github.com/kyroy/kdtree
RUN sed -i 's|../../..||g' config.yaml

CMD [ "go", "run", "main.go" ]
