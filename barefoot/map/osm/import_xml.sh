#!/bin/bash

#
# Copyright (C) 2015, BMW Car IT GmbH
#
# Author: Sebastian Mattheis <sebastian.mattheis@bmw-carit.de>
#
# Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
# in compliance with the License. You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0 Unless required by applicable law or agreed to in
# writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
# WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific
# language governing permissions and limitations under the License.
#

if [ "$#" -eq "6" ] && ( [ "$6" = "slim" ] || [ "$6" = "normal" ] )
then
	input=$1
	database=$2
	user=$3
	password=$4
	config=$5
	mode=$6
	seg_sa2s=$7
elif [ "$#" -eq "0" ]
then
	input=/mnt/map/osm/australia.osm
	database=australia
	user=osmuser
	password=pass
	config=/mnt/map/tools/road-types.json
	seg_sa2s=/mnt/map/tools/seg_sa2s.json
	mode=slim
else
	echo "Error. Say '$0 osm-file database user password bfmap-config slim|normal' or run with defaults '$0'."
	exit
fi

echo "Start creation and initialization of database '${database}' ..."
sudo -u postgres createdb ${database}
sudo -u postgres psql -d ${database} -c "CREATE EXTENSION hstore;"
sudo -u postgres psql -d ${database} -c "CREATE EXTENSION postgis;"
echo "Done."

echo "Start creation of user and initialization of credentials ..."
sudo -u postgres psql -c "CREATE USER \"${user}\" PASSWORD '${password}';"
sudo -u postgres psql -c "GRANT ALL ON DATABASE \"${database}\" TO \"${user}\";"
passphrase="localhost:5432:${database}:${user}:${password}"
if [ ! -e ~/.pgpass ] || [ `less ~/.pgpass | grep -c "$passphrase"` -eq 0 ]
then
	echo "$passphrase" >> ~/.pgpass
	chmod 0600 ~/.pgpass
fi
echo "Done."

echo "Start population of OSM data (ogr2ogr) ..."
psql -h localhost -d ${database} -U ${user} -f /mnt/map/osm/pgsnapshot_schema_0.6.sql
rm -rf /mnt/map/osm/tmp
mkdir /mnt/map/osm/tmp
if [ -z "$JAVACMD_OPTIONS" ]; then
    JAVACMD_OPTIONS="-Djava.io.tmpdir=/mnt/map/osm/tmp"
else
    JAVACMD_OPTIONS="$JAVACMD_OPTIONS -Djava.io.tmpdir=/mnt/map/osm/tmp"
fi
export JAVACMD_OPTIONS

#echo "ogr2ogr command starting"
#ogr2ogr -f "PostgreSQL" PG:"host=localhost user=${user} port=5432 dbname=${database} password=${password}" ${input} #-nln ways -append -skipfailures

echo "osmosis"
echo ${input}
/opt/osmosis/package/bin/osmosis --read-xml file=${input} --tf accept-ways highway=* route=ferry --write-pgsql user="${user}" password="${password}" database="${database}"

echo "Done."

#echo "Start extraction of routing data (bfmap tools) ..."
cd /mnt/map/tools/
if [ "$mode" = "slim" ]
then
	echo "one"
	python osm2ways.py --host localhost --port 5432 --database ${database} --table temp_ways --user ${user} --password ${password} --slim
elif [ "$mode" = "normal" ]
then
	echo "two"
	python osm2ways.py --host localhost --port 5432 --database ${database} --table temp_ways --user ${user} --password ${password} --prefix _tmp
fi
echo "three"
python ways2bfmap.py --source-host localhost --source-port 5432 --source-database ${database} --source-table temp_ways --source-user ${user} --source-password ${password} --target-host localhost --target-port 5432 --target-database ${database} --target-table bfmap_ways --target-user ${user} --target-password ${password} --config ${config} --seg_sa2s ${seg_sa2s}
echo "Done."
