# Yulo
Created Friday 27 April 2018

I am tentatively naming the processing framework (as an adaptable component built for the larger [Freight Performance and Measurement Project](./FPMP.md)) "Yulo". Yulo is the [Awabakal word for foot.](http://hri.newcastle.edu.au/AwabakalEnglish/Awabakal1834.html) This reflects both the debt owed to the [Barefoot](./Processing/Barefoot.md) repository and the [pre-European elements](http://hri.newcastle.edu.au/AwabakalEnglish/Awabakal1834.html) of the Australian road network including Watt St in Newcastle in Awabakal country.

Yulo as a framework consists of the following elements

Homogenisation of raw data including [read in](./Raw_data_and_perparation/Data_read_in.md) and adjustments of [timezones](./Raw_data_and_perparation/Timezones.md).
[Grouping GPS observations into trips and stops](./Processing/Tripgrouping.md) using a clustering mechanism.
Identifying [anonymous stop clusters](./Processing/SummaryStops.md).
Matching GPS coordinates to the road network and imputing the use of other roads using a modified version of [Barefoot](./Processing/Barefoot.md).
[Extracing data from barefoot JSON](./Processing/postbarefoot.md) and imputing other data
A database structure for [storing processed data](./Storage_of_processed_data.md) and [querying it](./Display_and_queries/Queries.md)

