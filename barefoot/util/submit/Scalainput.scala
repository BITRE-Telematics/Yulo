import BroadcastMatcher.BroadcastMatcher
// Instantiate map matcher as broadcast variable in Spark Context (sc).
val matcher = sc.broadcast(new BroadcastMatcher("localhost", 1234, "australia", "user", "pass", "/media/veracrypt12/barefoot/config"))

// Load trace data as RDD from CSV file asset of tuples:
// (object-id: String, time: Long, position: Point)
val traces = sc.textFile("traces.csv").map(x => {
  val y = x.split(",")
  (y(0), y(1).toLong, new Point(y(2).toDouble, y(3).toDouble))
})

// Run a map job on RDD that uses the matcher instance.
val matches = traces.groupBy(x => x._1).map(x => {
  val trip = x._2.map({
    x => new MatcherSample(x._1, x._2, x._3)
  }).toList
  matcher.mmatch(trip)
)
