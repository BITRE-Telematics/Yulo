# Timezones
Created Friday 27 April 2018

This script can be run after the read in on the resulting CSV. It matches each point to a state, transforms the unix epoch into time using the timezone specified in readin, converts that to the local time specified by the state, and converts back into a unix epoch. This is done using Olson timezones.

This means the datetime in the database will always be the local time, taking into account DST and other changes. This will also conflict with local use in place like Broken Hill and Eucla which use the neighbouring state's time.

