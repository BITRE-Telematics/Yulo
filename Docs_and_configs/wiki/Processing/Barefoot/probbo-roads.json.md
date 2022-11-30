# probbo-roads.json
Created Tuesday 01 May 2018

This is a files with specific road segments that are [causing problems](../Matching_issues_and_errors/GPS_error.md). I am unsure whether this will be used yet [It won't 2019-05-08]. It has a similar structure to [road-types.json](./road-types.json.md) and works with the same functions from [bfmap.py](./bfmap.py.md).
Each element is a way/segment that is providing particular problems and requires a specific priority weighting to push vehicles towards or away from it. We can only feel this out gradually. 
AFAICT the id field in either JSON file is not used in any script. Here I can used it simply to aid identification by the editor.
 

The current version has no roads in it as I am experiementing with a rerun capacity as part of [PostBarefoot.r.](../postbarefoot.md)

