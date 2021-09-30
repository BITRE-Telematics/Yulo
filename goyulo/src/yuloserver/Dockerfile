FROM golang:1.10
LABEL maintainer="Telematics@infrastructure.gov.au"


COPY ./ /go/src/yuloserver
WORKDIR /go/src/yuloserver
RUN chmod  777 -R /go/src/yuloserver

RUN apt-get update && \
 apt-get install -y libgeos-dev && \
 apt-get install -y netcat

RUN go get github.com/neo4j/neo4j-go-driver/neo4j && \
 go get github.com/paulmach/orb && \
 go get github.com/paulsmith/gogeos/geos && \
 go get github.com/tidwall/gjson && \
 go get github.com/xiam/to && \
 go get github.com/xiam/dig && \
 go get github.com/kyroy/kdtree && \
 go get github.com/shirou/gopsutil/cpu
RUN ls /go/src/
RUN sed -i 's|../../..||g' config.yaml
#ENTRYPOINT ["/go/src/yuloserver"]
CMD [ "go", "run", "/go/src/yuloserver/main.go" ]
~
