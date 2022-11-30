# rest and loading areas.py
Created Wednesday 24 July 2019

This script uploads data on rest areas and loading zones. Most of this data is taken from data.gov.au and is usually downloaded in script.
Attribute names as usually whatever NSW used.

Queensland data needed to be downloaded manually.
Victorian rest area data was not open but the URL to the csv was publically available in the javascript for a website displaying the data.
Tasmanian data was available on request and was rendered in an idiosyncratic projection.
NT data was scraped from a table and manually adjusted.
WA data was missing lat lon fields in the official download but the url of a correct csv was available from the data website's javascript.
The ACT's one (1) rest area was created manually.
SA loading zones had a malformed csv and the bespoke read in fixes this.

