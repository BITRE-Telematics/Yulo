# Storage of pre processed data
Created Friday 27 April 2018

Raw data
The data provided by firms comes in a wide variety of formats, including csv, xlsx, xlxb, xml and json. Apart from the select number of fields that are effectively homogenous, there are many fields that are distinct to a given firm. We are unwilling to discard this data lest we find a use for it in future.
However there are no good options for a database that stores multiple file formats. Fortunately, all provided formats thus far can be represented in a flat database – that is in a table structure. As such, I suggest the raw data be stored in a database with a distinct table for each firm, containing the fields as described in that firm’s raw data.

Homogenised data
We may wish to also store the data in a table after it has been homogenised (reduced to common fields) but before processing. This is probably made crucial by the need to store residual data from the trip grouping algorithm to be processed in the next month.

