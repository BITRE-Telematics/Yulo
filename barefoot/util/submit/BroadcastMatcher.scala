object BroadcastMatcher {
  private var instance = null: Matcher

  private def initialize(host: String, port: Int, name: String, user: String, pass: String, config: String) {
    if (instance != null) return
    this.synchronized {
      if (instance == null) { // initialize map matcher once per Executor (JVM process/cluster node)
        val reader = new PostGISReader(host, port, name, "bfmap_ways", user, pass, Configuration.read(new JSONObject(config)))
        val map = RoadMap.Load(reader)

        map.construct();

        val router = new Dijkstra[Road, RoadPoint]()
        val cost = new TimePriority()
        val spatial = new Geography()

        instance = new Matcher(map, router, cost, spatial)
      }
    }
  }
}

@SerialVersionUID(1L)
class BroadcastMatcher(host: String, port: Int, name: String, user: String, pass: String, config: String) extends Serializable {

  def mmatch(samples: List[MatcherSample]): MatcherKState = {
    mmatch(samples, 0, 0)
  }

  def mmatch(samples: List[MatcherSample], minDistance: Double, minInterval: Int): MatcherKState = {
    BroadcastMatcher.initialize(host, port, name, user, pass, config)
    BroadcastMatcher.instance.mmatch(new ArrayList[MatcherSample](samples.asJava), minDistance, minInterval)
  }
}
