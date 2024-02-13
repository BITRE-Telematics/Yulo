#include <Rcpp.h>

using namespace Rcpp;

// This is a simple example of exporting a C++ function to R. You can
// source this function into an R session using the Rcpp::sourceCpp 
// function (or via the Source button on the editor toolbar). Learn
// more about Rcpp at:
//
//   http://www.rcpp.org/
//   http://adv-r.had.co.nz/Rcpp.html
//   http://gallery.rcpp.org/
//

constexpr uint64_t kInvalidGraphId = 0x3fffffffffff;
constexpr uint32_t kMaxGraphHierarchy = 7;
constexpr uint32_t kMaxGraphTileId = 4194303;
constexpr uint32_t kMaxGraphId = 2097151;
constexpr uint64_t kIdIncrement = 1 << 25;

// Road class or importance of an edge
enum class RoadClass : uint8_t {
  kMotorway = 0,
    kTrunk = 1,
    kPrimary = 2,
    kSecondary = 3,
    kTertiary = 4,
    kUnclassified = 5,
    kResidential = 6,
    kServiceOther = 7
};


struct TileLevel {
  uint8_t level;
  RoadClass importance;
  std::string name;
  midgard::Tiles<midgard::PointLL> tiles;
};

struct GraphId {
public:
  // Single 64 bit value representing the graph id.
  // Bit fields within the Id include:
  //      3  bits for hierarchy level
  //      22 bits for tile Id (supports lat,lon tiles down to 1/8 degree)
  //      21 bits for id within the tile.
  uint64_t value;
  
  /**
   * Default constructor
   */
  GraphId() : value(kInvalidGraphId) {
  }
  
  /**
   * Constructor.
   * @param  tileid Tile Id.
   * @param  level  Hierarchy level
   * @param  id     Unique identifier within the level. Cast this to 64 bits
   *                since the Id portion of the value crosses the 4-byte bdry.
   */
  GraphId(const uint32_t tileid, const uint32_t level, const uint32_t id) {
    if (tileid > kMaxGraphTileId) {
      throw std::logic_error("Tile id out of valid range");
    }
    if (level > kMaxGraphHierarchy) {
      throw std::logic_error("Level out of valid range");
    }
    if (id > kMaxGraphId) {
      throw std::logic_error("Id out of valid range");
    }
    value = level | (tileid << 3) | (static_cast<uint64_t>(id) << 25);
  }
  
  /**
   * Constructor
   * @param value all the various bits rolled into one
   */
  explicit GraphId(const uint64_t value) : value(value) {
    if (tileid() > kMaxGraphTileId) {
      throw std::logic_error("Tile id out of valid range");
    }
    if (level() > kMaxGraphHierarchy) {
      throw std::logic_error("Level out of valid range");
    }
    if (id() > kMaxGraphId) {
      throw std::logic_error("Id out of valid range");
    }
  }
  
  /**
   * Constructor
   * @param value a string of the form level/tile_id/id
   */
  explicit GraphId(const std::string& value) {
    std::vector<uint32_t> values;
    std::string::size_type pos = 0;
    while (pos != std::string::npos) {
      auto next = value.find('/', pos + (pos > 0));
      values.push_back(std::stoul(value.substr(pos + (pos > 0), next)));
      pos = next;
    }
    if (values.size() != 3)
      throw std::logic_error("Tile string format does not match level/tile/id");
    *this = GraphId(values[1], values[0], values[2]);
  }
  
  /**
   * Gets the tile Id.
   * @return   Returns the tile Id.
   */
  inline uint32_t tileid() const {
    return (value & 0x1fffff8) >> 3;
  }
  
  /**
   * Gets the hierarchy level.
   * @return   Returns the level.
   */
  inline uint32_t level() const {
    return (value & 0x7);
  }
  
  /**
   * Gets the identifier within the hierarchy level.
   * @return   Returns the unique identifier within the level.
   */
  inline uint32_t id() const {
    return (value & 0x3ffffe000000) >> 25;
  }
  
  /**
   * Set the Id portion of the GraphId. Since the Id crosses the 4-byte
   * boundary cast it to 64 bits.
   * @param  id  Id to set.
   */
  void set_id(const uint32_t id) {
    value = (value & 0x1ffffff) | (static_cast<uint64_t>(id & 0x1fffff) << 25);
  }
  
  /**
   * Conversion to bool for use in conditional statements.
   * Note that this is explicit to avoid unexpected implicit conversions. Some
   * statements, including "if", "&&", "||", "!" are "implicit explicit" and
   * will result in conversion.
   * @return boolean true if the id is valid.
   */
  explicit inline operator bool() const {
    return Is_Valid();
  }
  
  /**
   * Returns true if the id is valid
   * @return boolean true if the id is valid
   */
  bool Is_Valid() const {
    // TODO: make this strict it should check the tile hierarchy not bit field widths
    return value != kInvalidGraphId;
  }
  
  /**
   * Returns a GraphId omitting the id of the of the object within the level.
   * Construct a new GraphId with the Id portion omitted.
   * @return graphid with only tileid and level included
   */
  GraphId Tile_Base() const {
    return GraphId((value & 0x1ffffff));
  }
  
  /**
   * Returns a value indicating the tile (level and tile id) of the graph Id.
   * @return  Returns a 32 bit value.
   */
  inline uint32_t tile_value() const {
    return (value & 0x1ffffff);
  }
  

  /**
   * Post increments the id.
   */
  GraphId operator++(int) {
    GraphId t = *this;
    value += kIdIncrement;
    return t;
  }
  
  /**
   * Pre increments the id.
   */
  GraphId& operator++() {
    value += kIdIncrement;
    return *this;
  }
  
  /**
   * Advances the id
   */
  GraphId operator+(uint64_t offset) const {
    return GraphId(tileid(), level(), id() + offset);
  }
  
  /**
   * Less than operator for sorting.
   * @param  rhs  Right hand side graph Id for comparison.
   * @return  Returns true if this GraphId is less than the right hand side.
   */
  bool operator<(const GraphId& rhs) const {
    return value < rhs.value;
  }
  
  // Operator EqualTo.
  bool operator==(const GraphId& rhs) const {
    return value == rhs.value;
  }
  
  // Operator not equal
  bool operator!=(const GraphId& rhs) const {
    return value != rhs.value;
  }
  
  // cast operator
  operator uint64_t() const {
    return value;
  }
  
  // Stream output
  friend std::ostream& operator<<(std::ostream& os, const GraphId& id);
};

const std::vector<TileLevel>& levels() {
  // Static tile levels
  static const std::vector<TileLevel> levels_ = {
    
    TileLevel{0, stringToRoadClass("Primary"), "highway",
              midgard::Tiles<midgard::PointLL>{{{-180, -90}, {180, 90}},
                                               4,
                                               static_cast<unsigned short>(kBinsDim)}},
                                               
                                               TileLevel{1, stringToRoadClass("Tertiary"), "arterial",
                                                         midgard::Tiles<midgard::PointLL>{{{-180, -90}, {180, 90}},
                                                                                          1,
                                                                                          static_cast<unsigned short>(kBinsDim)}},
                                                                                          
                                                                                          TileLevel{2, stringToRoadClass("ServiceOther"), "local",
                                                                                                    midgard::Tiles<midgard::PointLL>{{{-180, -90}, {180, 90}},
                                                                                                                                     .25,
                                                                                                    static_cast<unsigned short>(kBinsDim)}},
  };
  
  return levels_;
}



const TileLevel& GetTransitLevel() {
  // Should we make a class lower than service other for transit?
  static const TileLevel transit_level_ =
    {3, stringToRoadClass("ServiceOther"), "transit",
     midgard::Tiles<midgard::PointLL>{{{-180, -90}, {180, 90}},
                                      .25,
     static_cast<unsigned short>(kBinsDim)}};
  
  return transit_level_;
}


std::string FileSuffix(const GraphId& graphid,
                                  const std::string& fname_suffix,
                                  bool is_file_path,
                                  const TileLevel* tiles) {
  /*
   if you have a graphid where level == 8 and tileid == 24134109851 you should get:
   8/024/134/109/851.gph since the number of levels is likely to be very small this limits the total
   number of objects in any one directory to 1000 which is an empirically derived good choice for
   mechanical hard drives this should be fine for s3 as well (even though it breaks the rule of most
   unique part of filename first) because there will be just so few objects in general in practice
   */
  
  // figure the largest id for this level
  if ((tiles && tiles->level != graphid.level()) ||
      (!tiles && graphid.level() >= levels().size() &&
      graphid.level() != GetTransitLevel().level)) {
    throw std::runtime_error("Could not compute FileSuffix for GraphId with invalid level: " +
                             std::to_string(graphid));
  }
  
  // get the level info
  const auto& level = tiles ? *tiles
  : (graphid.level() == GetTransitLevel().level
       ? GetTransitLevel()
         : levels()[graphid.level()]);
  
  // figure out how many digits in tile-id
  const uint32_t max_id = static_cast<uint32_t>(level.tiles.ncolumns() * level.tiles.nrows() - 1);
  
  if (graphid.tileid() > max_id) {
    throw std::runtime_error("Could not compute FileSuffix for GraphId with invalid tile id:" +
                             std::to_string(graphid));
  }
  size_t max_length = static_cast<size_t>(std::log10(std::max(1u, max_id))) + 1;
  const size_t remainder = max_length % 3;
  if (remainder) {
    max_length += 3 - remainder;
  }
  assert(max_length % 3 == 0);
  
  // Calculate tile-id string length with separators
  const size_t tile_id_strlen = max_length + max_length / 3;
  assert(tile_id_strlen % 4 == 0);
  
  const char separator =  '/';
  
  std::string tile_id_str(tile_id_strlen, '0');
  size_t ind = tile_id_strlen - 1;
  for (uint32_t tile_id = graphid.tileid(); tile_id != 0; tile_id /= 10) {
    tile_id_str[ind--] = '0' + static_cast<char>(tile_id % 10);
    if ((tile_id_strlen - ind) % 4 == 0) {
      ind--; // skip an additional character to leave space for separators
    }
  }
  // add separators
  for (size_t sep_ind = 0; sep_ind < tile_id_strlen; sep_ind += 4) {
    tile_id_str[sep_ind] = separator;
  }
  
  return std::to_string(graphid.level()) + tile_id_str + fname_suffix;
}



// [[Rcpp::export('get_dir')]]
String handle_get_traffic_dir(uint64_t way_id) {
  GraphId graph_id(way_id);
  auto tile_path = FileSuffix(graph_id);

  auto dir_str = tile_path.string();

  
  
  return dir_str;
}
