# Further analysis
Created Thursday 31 May 2018

Beyond the particular objectives of the project, the data produced lends itself to some other analysis including
[community detection](./Further_analysis/community_detection.md)

I have put some thought towards machine learning, including neural networks etc mainly because we will likely be asked about it. There's no clear case where the tools will be useful in this case. That is, we don't have a huge body of classified data and an opportunity where classifying as yet unclassified data will be useful.
For instance, we could develop an algorithm that could predict where a truck will go based on the beginning of it's journey but, quite apart from the veractiy of the results, what would be the point? They know where they are going, and we gain nothing by knowing in advance.
One case that might be useful is imputing a truck's type when it is unknown by training the data on the movements of trucks we do no, so we can have deeper information on the behaviour of different types. The data may be insufficient for this, and there will be huge problems with disentangling factors correlated to firms rather than truck type per se.
Another is creating an adjusted activity/performance measure where the adjustment controls for a whole basket of different exogenous effects (weather, time of day/year etc) that can have complex and non linear interactions on performance. This would be similar to [this by ABARES](http://www.agriculture.gov.au/abares/research-topics/climate/farm-performance-climate). However the main reason would be because we can, not necessarily because it's a good use of resources.

