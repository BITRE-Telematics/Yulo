FROM ubuntu
ENV DEBIAN_FRONTEND=noninteractive
COPY requirements.txt ./
RUN apt update
RUN apt-get install python3.6 -y
RUN apt-get install python3-pip -y
RUN apt-get install libspatialindex-dev gdal-bin libgdal-dev -y
RUN pip install --no-cache-dir -r requirements.txt
VOLUME ['/script']
WORKDIR /script
