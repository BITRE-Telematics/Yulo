#!/bin/bash
set -e
#updating repository data
#Docker
apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common


curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"

#GDAL
add-apt-repository -y ppa:ubuntugis/ppa


echo "deb http://cran.rstudio.com/bin/linux/ubuntu xenial/" | sudo tee -a /etc/apt/sources.list
gpg --keyserver keyserver.ubuntu.com --recv-key E084DAB9
gpg -a --export E084DAB9 | sudo apt-key add -

#neo4j
wget -O - https://debian.neo4j.org/neotechnology.gpg.key | apt-key add -
echo 'deb https://debian.neo4j.org/repo stable/' | tee -a /etc/apt/sources.list.d/neo4j.list


apt-get update

##barefoot requirements
apt-get install maven
apt-get install openjdk-8-jdk
apt-get install docker-ce


apt install build-essential git cmake pkg-config \
libbz2-dev libxml2-dev libzip-dev libboost-all-dev \
lua5.2 liblua5.2-dev libtbb-dev

##misc dependencies

apt-get install libgeos-dev
apt-get install libcurl
apt-get install libv8-3.14-dev
apt-get install libssl
apt-get install libxml2-dev
apt-get install cargo
apt-get install gdal-bin
apt-get install libgdal-dev

apt-get install lidudunits-dev

apt-get install pip3
python3 -m pip install --upgrade pip

##These can be put in a requirements.txt for pip installation

pip3 install numpy
pip3 install pandas
pip3 install geopy
pip3 install numba
pip3 install neo4j
pip3 install neobolt
pip3 install geopandas
pip3 install sklearn
pip3 install shapely
pip3 install awscli
apt install python3-rtree

#untested, you can also use the docker file
apt-get install golang-go
# #this doesn't work, the main problem is the yaml package which keeps calling on repositories that are no longer there
# ##it can be manually copied to the the GOPATH or fix module management
# echo "changing go path"
# echo export GOPATH=../goyulo >> ~/.profile
# source ~/.profile
# #




go get github.com/neo4j/neo4j-go-driver/neo4j
go get github.com/paulmach/orb
go get github.com/paulsmith/gogeos/geos
go get github.com/tidwall/gjson
go get github.com/gosexy/to
go get github.com/gosexy/dig
go get github.com/kyroy/kdtree
##move goyulo/src/yuloserver/yaml to GOPATH folder if things keep failing to work

##adding capacity for more open files
##echo  ulimit -n 56636 >> ~/.profile
##echo  ulimit -u 56636 >> ~/.profile
##source ~/.profile
echo  fs.file-max =  56636 >> /etc/sysctl.conf
sysctl -p

## download shapefiles and transform for goyulo and for other processes
bash shapefiles.sh

##R stuff for Shinyleaflet and analysis
apt-get install r-base
RScript Rinstallations.R

## Local neo4j installation if used
#apt-get install neo4j
#cp ../Docs_and_configs/neo4j.conf /etc/neo4j/neo4j.conf
#cp ../Docs_and_configs/neo4j /etc/default/neo4j

#curl https://github.com/neo4j-contrib/neo4j-apoc-procedures/releases/download/3.5.0.2/apoc-3.5.0.2-all.jar -o /var/lib/neo4j/plugins/apoc-3.5.0.2-all.jar

#curl https://github.com/neo4j-contrib/spatial/releases/download/0.26.2-neo4j-3.5.2/neo4j-spatial-0.26.2-neo4j-3.5.2-server-plugin.jar -o /var/lib/neo4j/plugins/neo4j-spatial-0.26.2-neo4j-3.5.2-server-plugin.jar



##osrm
##cd ../osrm
##git clone https://github.com/Project-OSRM/osrm-backend.git
##cd osrm-backend
##mkdir -p build
##cd build
##cmake .. -DCMAKE_BUILD_TYPE=Release
##cmake --build .
##sudo cmake --build . --target install
## mv osrm ../../
##cd ../..
##rm osrm-backend