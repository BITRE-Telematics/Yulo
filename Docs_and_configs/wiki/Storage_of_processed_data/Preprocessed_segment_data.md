# Preprocessed segment data
Created Friday 29 November 2019

Given the time it takes to query segments with many observations fro the database for breakdowns by hours etc. I am moving to regular preprocessing for trips, vehicles and imputed speed quartiles which will be stored as arrays on the segment. These won't be as up to date as new queries but will be much faster and sufficient for most users. See the [schema](./schema.md) for more details

