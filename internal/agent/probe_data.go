package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"nw-guardian/internal"
	"strings"
	"time"
)

/*

storing the agent probe types will be more complicated



*/

type ProbeData struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	ProbeID   primitive.ObjectID `json:"probe" bson:"probe"`
	Triggered bool               `json:"triggered" bson:"triggered"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
	/*
		when we are storing agent probe data we will use the agent as the one reporting,
		and the target will be the targeted agent / host
	*/
	Target ProbeTarget `bson:"target" json:"target"`
	Data   interface{} `json:"data,omitempty" bson:"data,omitempty"`
}

func DeleteProbeDataByProbeID(db *mongo.Database, probeID primitive.ObjectID) error {
	ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "probe_data.DeleteProbeDataByProbeID", ObjectID: probeID}

	// Convert the string ID to an ObjectID
	// Create a filter to match the document by ID
	filter := bson.M{"probe": probeID}

	// Perform the deletion
	_, err := db.Collection("probe_data").DeleteMany(context.TODO(), filter)
	if err != nil {
		ee.Message = "unable to delete probe data by id"
		ee.Error = err
		return ee.ToError()
	}

	return nil
}

func (pd *ProbeData) Create(db *mongo.Database) error {
	// todo handle to check if agent id is set and all that... or should it be in the api section??
	ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "probe_data.Create", ObjectID: pd.ProbeID}

	pd.ID = primitive.NewObjectID()

	pp := Probe{ID: pd.ProbeID}
	pp2, err := pp.Get(db)
	if err != nil {
		ee.Message = "no matching probe found"
		ee.Error = err
		return ee.ToError()
	}

	if len(pp2) < 1 {
		log.Printf("pp: %v", pp)
		log.Printf("pp2: %v", pp2)
		return errors.New("no matching probe found")
	}

	a := Agent{ID: pp2[0].Agent}
	err = a.UpdateTimestamp(db)
	if err != nil {
		ee.Message = "couldnt update timestamp on agent"
		ee.Error = err
		ee.Level = log.WarnLevel
		ee.Print()
	}

	// load types

	pd.Data, _ = pd.parse(db)

	if (pd.CreatedAt == time.Time{}) {
		pd.CreatedAt = time.Now()
		// don't return??!??
		ee.Message = "timestamp not included in probe data"
		ee.Error = err
		ee.Level = log.WarnLevel
		ee.Print()
	}

	mar, err := bson.Marshal(pd)
	if err != nil {
		ee.Message = "unable to marshal"
		ee.Error = err
		return ee.ToError()
	}

	var b *bson.D
	err = bson.Unmarshal(mar, &b)
	if err != nil {
		ee.Message = "unable to unmarshal"
		ee.Error = err
		return ee.ToError()
	}
	_, err = db.Collection("probe_data").InsertOne(context.TODO(), b)
	if err != nil {
		ee.Message = "error inserting doc"
		ee.Error = err
		return ee.ToError()
	}
	return nil
}

func (pd *ProbeData) parse(db *mongo.Database) (interface{}, error) {
	ee := internal.ErrorFormat{Package: "internal.agent", Level: log.WarnLevel, Function: "probe_data.parse", ObjectID: pd.ProbeID}

	pp := Probe{ID: pd.ProbeID}
	probes, err := pp.Get(db)
	if err != nil || len(probes) == 0 {
		ee.Message = "unable to get probe from id or no probes found"
		ee.Error = err
		return nil, ee.ToError()
	}

	probe := probes[0]
	probeType := probe.Type

	// If it's an AGENT type, try parsing actual type from ProbeTarget
	if probeType == ProbeType_AGENT {
		parts := strings.Split(pd.Target.Target, "%%%")
		if len(parts) >= 1 {
			probeType = ProbeType(parts[0])
		}
	}

	jsonData, err := json.Marshal(pd.Data)
	if err != nil {
		ee.Message = "cannot marshal"
		ee.Error = err
		return nil, ee.ToError()
	}

	switch probeType {
	case ProbeType_TRAFFICSIM:
		var stats TrafficSimClientStats
		err = json.Unmarshal(jsonData, &stats)
		return stats, err

	case ProbeType_RPERF:
		var rperf RPerfResults
		err = json.Unmarshal(jsonData, &rperf)
		return rperf, err

	case ProbeType_MTR:
		var mtr MtrResult
		err = json.Unmarshal(jsonData, &mtr)
		return mtr, err

	case ProbeType_NETWORKINFO:
		var netinfo NetResult
		err = json.Unmarshal(jsonData, &netinfo)
		return netinfo, err

	case ProbeType_PING:
		var ping PingResult
		err = json.Unmarshal(jsonData, &ping)
		return ping, err

	case ProbeType_SPEEDTEST:
		var result SpeedTestResult
		err = json.Unmarshal(jsonData, &result)
		if err == nil {
			_ = pp.UpdateFirstProbeTarget(db, "ok")
		}
		return result, err

	case ProbeType_SPEEDTEST_SERVERS:
		var servers []SpeedTestServer
		err = json.Unmarshal(jsonData, &servers)
		return servers, err

	case ProbeType_SYSTEMINFO:
		var sysinfo CompleteSystemInfo
		err = json.Unmarshal(jsonData, &sysinfo)
		return sysinfo, err

	default:
		ee.Message = fmt.Sprintf("unsupported probe type: %s", probeType)
		return nil, ee.ToError()
	}
}

// GroupedProbeData represents probe data grouped by reporting agent, target agent, and type
type GroupedProbeData struct {
	ReportingAgent primitive.ObjectID `json:"reportingAgent"`
	TargetAgent    primitive.ObjectID `json:"targetAgent"`
	ProbeType      string             `json:"probeType"`
	Data           []ProbeData        `json:"data"`
}

// AgentGroupedData represents all data grouped by agents and types
type AgentGroupedData struct {
	// Map structure: ReportingAgent -> TargetAgent -> ProbeType -> []ProbeData
	Groups map[string]map[string]map[string][]ProbeData `json:"groups"`

	// Available agents from the probe configuration
	AvailableTargets []ProbeTarget `json:"availableTargets"`

	// Summary information
	Summary GroupingSummary `json:"summary"`
}

// GroupingSummary provides summary statistics
type GroupingSummary struct {
	TotalDataPoints int                  `json:"totalDataPoints"`
	ReportingAgents []primitive.ObjectID `json:"reportingAgents"`
	TargetAgents    []primitive.ObjectID `json:"targetAgents"`
	ProbeTypes      []string             `json:"probeTypes"`
	DataCountByType map[string]int       `json:"dataCountByType"`
}

// GetAgentProbeDataGrouped retrieves and groups probe data by reporting agent, target agent, and type
func (probe *Probe) GetAgentProbeDataGrouped(req *ProbeDataRequest, db *mongo.Database) (*AgentGroupedData, error) {
	ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "probe_data.GetAgentProbeDataGrouped"}

	// First, get the probe configuration to extract available targets
	probeConfig, err := probe.Get(db)
	if err != nil || len(probeConfig) == 0 {
		ee.Error = errors.New("could not find probe configuration")
		return nil, ee.ToError()
	}

	availableTargets := probeConfig[0].Config.Target

	// Build base filter
	filter := bson.M{
		"probe": probe.ID,
	}

	// Add agent filter if specified
	if probe.Agent != primitive.NilObjectID {
		filter["agent"] = probe.Agent
	}

	// For time filtering, we'll need to handle different timestamp fields
	// We'll use a general approach and let MongoDB handle non-existent fields
	if !req.Recent {
		timeFilter := bson.M{
			"$or": []bson.M{
				{"data.stop_timestamp": bson.M{"$gt": req.StartTimestamp, "$lt": req.EndTimestamp}},
				{"data.timestamp": bson.M{"$gt": req.StartTimestamp, "$lt": req.EndTimestamp}},
				{"data.reportTime": bson.M{"$gt": req.StartTimestamp, "$lt": req.EndTimestamp}},
			},
		}
		filter["$and"] = []bson.M{timeFilter}
	}

	// Set query options - we'll sort by _id descending as a general approach
	opts := options.Find().
		SetLimit(req.Limit).
		SetSort(bson.D{{"_id", -1}})

	// Execute query
	cursor, err := db.Collection("probe_data").Find(context.TODO(), filter, opts)
	if err != nil {
		ee.Message = "cannot find probe data"
		ee.Error = err
		return nil, ee.ToError()
	}
	defer cursor.Close(context.TODO())

	// Initialize the result structure
	result := &AgentGroupedData{
		Groups:           make(map[string]map[string]map[string][]ProbeData),
		AvailableTargets: availableTargets,
		Summary: GroupingSummary{
			DataCountByType: make(map[string]int),
			ReportingAgents: []primitive.ObjectID{},
			TargetAgents:    []primitive.ObjectID{},
			ProbeTypes:      []string{},
		},
	}

	// Maps to track unique agents and types
	reportingAgentsMap := make(map[string]primitive.ObjectID)
	targetAgentsMap := make(map[string]primitive.ObjectID)
	probeTypesMap := make(map[string]bool)

	// Process results
	var probeDataList []ProbeData
	if err := cursor.All(context.TODO(), &probeDataList); err != nil {
		ee.Message = "error decoding probe data"
		ee.Error = err
		return nil, ee.ToError()
	}

	// Group the data
	for _, data := range probeDataList {
		// Extract reporting agent (group) and target agent
		reportingAgentID := data.Target.Group // The agent sending the data
		targetAgentID := data.Target.Agent    // The agent being monitored

		// Skip if no valid agents
		if reportingAgentID == primitive.NilObjectID || targetAgentID == primitive.NilObjectID {
			continue
		}

		// Extract probe type from target string
		probeType := extractProbeType(data.Target.Target)
		if probeType == "" {
			log.Warnf("Could not extract probe type from target: %s", data.Target.Target)
			continue
		}

		reportingKey := reportingAgentID.Hex()
		targetKey := targetAgentID.Hex()

		// Track unique agents and types
		reportingAgentsMap[reportingKey] = reportingAgentID
		targetAgentsMap[targetKey] = targetAgentID
		probeTypesMap[probeType] = true

		// Initialize nested maps if needed
		if _, exists := result.Groups[reportingKey]; !exists {
			result.Groups[reportingKey] = make(map[string]map[string][]ProbeData)
		}
		if _, exists := result.Groups[reportingKey][targetKey]; !exists {
			result.Groups[reportingKey][targetKey] = make(map[string][]ProbeData)
		}

		// Append data to the appropriate group
		result.Groups[reportingKey][targetKey][probeType] = append(
			result.Groups[reportingKey][targetKey][probeType],
			data,
		)

		// Update counters
		result.Summary.TotalDataPoints++
		result.Summary.DataCountByType[probeType]++
	}

	// Convert maps to slices for summary
	for _, agent := range reportingAgentsMap {
		result.Summary.ReportingAgents = append(result.Summary.ReportingAgents, agent)
	}
	for _, agent := range targetAgentsMap {
		result.Summary.TargetAgents = append(result.Summary.TargetAgents, agent)
	}
	for probeType := range probeTypesMap {
		result.Summary.ProbeTypes = append(result.Summary.ProbeTypes, probeType)
	}

	if result.Summary.TotalDataPoints == 0 {
		ee.Error = errors.New("no data found for this probe")
		return nil, ee.ToError()
	}

	return result, nil
}

// extractProbeType extracts the probe type from the target string
// Format: "PROBETYPE%%%actual_target" (e.g., "TRAFFICSIM%%%216.138.253.125:8677")
func extractProbeType(targetString string) string {
	parts := strings.Split(targetString, "%%%")
	if len(parts) >= 2 {
		return parts[0]
	}
	// If no separator found, try to infer from the format
	if strings.Contains(targetString, ":") && strings.Count(targetString, ".") == 3 {
		// Likely an IP:port format, could be TRAFFICSIM
		return "TRAFFICSIM"
	}
	return ""
}

// GetAgentProbeDataFlat returns a flattened view of grouped data
func (probe *Probe) GetAgentProbeDataFlat(req *ProbeDataRequest, db *mongo.Database) ([]GroupedProbeData, error) {
	// Get the grouped data first
	groupedData, err := probe.GetAgentProbeDataGrouped(req, db)
	if err != nil {
		return nil, err
	}

	// Flatten the nested structure
	var flatData []GroupedProbeData

	for reportingAgentHex, targetMap := range groupedData.Groups {
		reportingAgent, _ := primitive.ObjectIDFromHex(reportingAgentHex)

		for targetAgentHex, typeMap := range targetMap {
			targetAgent, _ := primitive.ObjectIDFromHex(targetAgentHex)

			for probeType, dataList := range typeMap {
				flatData = append(flatData, GroupedProbeData{
					ReportingAgent: reportingAgent,
					TargetAgent:    targetAgent,
					ProbeType:      probeType,
					Data:           dataList,
				})
			}
		}
	}

	return flatData, nil
}

// Helper function to get timestamp field based on probe type
func getTimestampField(probeType string) string {
	switch probeType {
	case "NETWORKINFO", "NETINFO", "SPEEDTEST", "SYSTEMINFO", "SYSINFO":
		return "data.timestamp"
	case "TRAFFICSIM":
		return "data.reportTime"
	default:
		return "data.stop_timestamp"
	}
}

// GetProbeTargetPairs extracts all unique reporting-target agent pairs from probe data
func GetProbeTargetPairs(probeID primitive.ObjectID, db *mongo.Database) ([]AgentPair, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"probe": probeID,
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"reporting": "$target.group",
					"target":    "$target.agent",
				},
			},
		},
		{
			"$project": bson.M{
				"_id":       0,
				"reporting": "$_id.reporting",
				"target":    "$_id.target",
			},
		},
	}

	cursor, err := db.Collection("probe_data").Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var pairs []AgentPair
	if err := cursor.All(context.TODO(), &pairs); err != nil {
		return nil, err
	}

	return pairs, nil
}

// GetProbeTypesFromData extracts all unique probe types from probe data
func GetProbeTypesFromData(probeID primitive.ObjectID, db *mongo.Database) ([]string, error) {
	// First get sample documents to extract probe types
	filter := bson.M{"probe": probeID}
	opts := options.Find().SetLimit(100) // Sample size

	cursor, err := db.Collection("probe_data").Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	probeTypesMap := make(map[string]bool)

	for cursor.Next(context.TODO()) {
		var data ProbeData
		if err := cursor.Decode(&data); err != nil {
			continue
		}

		probeType := extractProbeType(data.Target.Target)
		if probeType != "" {
			probeTypesMap[probeType] = true
		}
	}

	// Convert map to slice
	probeTypes := make([]string, 0, len(probeTypesMap))
	for pt := range probeTypesMap {
		probeTypes = append(probeTypes, pt)
	}

	return probeTypes, nil
}

// AgentPair represents a reporting-target agent pair
type AgentPair struct {
	ReportingAgent primitive.ObjectID `bson:"reporting" json:"reportingAgent"`
	TargetAgent    primitive.ObjectID `bson:"target" json:"targetAgent"`
}

// Example usage and integration with original GetData method
func (probe *Probe) GetAgentProbeData(req *ProbeDataRequest, db *mongo.Database) (map[string][]ProbeData, error) {
	// Get the grouped data
	groupedData, err := probe.GetAgentProbeDataGrouped(req, db)
	if err != nil {
		return nil, err
	}

	// Convert to simple map format if needed
	result := make(map[string][]ProbeData)

	for _, targetMap := range groupedData.Groups {
		for _, typeMap := range targetMap {
			for probeType, dataList := range typeMap {
				if existing, ok := result[probeType]; ok {
					result[probeType] = append(existing, dataList...)
				} else {
					result[probeType] = dataList
				}
			}
		}
	}

	return result, nil
}

// Example routes:
/*
tempRoutes = append(tempRoutes, &Route{
	Name: "Get Grouped Agent Probe Data",
	Path: "/probes/data/{probe}/grouped",
	JWT:  true,
	Func: func(ctx iris.Context) error {
		return GetAgentProbeDataGroupedRoute(ctx, r)
	},
	Type: RouteType_POST,
})

// You can also modify your existing route to use the new grouped functionality:
tempRoutes = append(tempRoutes, &Route{
	Name: "Get Probe Data",
	Path: "/probes/data/{probe}",
	JWT:  true,
	Func: func(ctx iris.Context) error {
		// Check if grouping is requested
		if ctx.URLParam("grouped") == "true" {
			return GetAgentProbeDataGroupedRoute(ctx, r)
		}
		// Otherwise use original handler
		// ... original code ...
	},
	Type: RouteType_POST,
})
*/

// GetData requires a checkrequest to be sent, if agent id is set,
// it will require the type to be sent in check, otherwise
// the check id will be used
func (probe *Probe) GetData(req *ProbeDataRequest, db *mongo.Database) ([]ProbeData, error) {
	ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "probe_data.GetData"}

	opts := options.Find().SetLimit(req.Limit)

	// Combined filter
	var combinedFilter = bson.M{"probe": probe.ID}
	if probe.Agent != (primitive.ObjectID{0}) {
		combinedFilter["agent"] = probe.Agent
		//combinedFilter["type"] = c.Type
		ee.ObjectID = probe.Agent
	}

	var timestampField = "data.stop_timestamp"

	if probe.Type == ProbeType_NETWORKINFO || probe.Type == ProbeType_SPEEDTEST || probe.Type == ProbeType_SYSTEMINFO {
		timestampField = "data.timestamp"
	} else if probe.Type == ProbeType_TRAFFICSIM {
		timestampField = "data.reportTime"
	}

	if !req.Recent {
		opts.SetSort(bson.D{{timestampField, -1}})
		timeFilter := bson.M{
			timestampField: bson.M{
				"$gt": req.StartTimestamp,
				"$lt": req.EndTimestamp,
			},
		}
		for k, v := range timeFilter {
			combinedFilter[k] = v
		}
	} else {
		opts = opts.SetSort(bson.D{{timestampField, -1}})
	}

	cursor, err := db.Collection("probe_data").Find(context.TODO(), combinedFilter, opts)
	if err != nil {
		ee.Message = "cannot find probed data"
		ee.Error = err
		return nil, ee.ToError()
	}

	var results []bson.D
	if err := cursor.All(context.TODO(), &results); err != nil {
		ee.Message = "error cursoring results"
		ee.Error = err
		return nil, ee.ToError()
	}

	if len(results) == 0 {
		ee.Error = errors.New("no data matches the provided check id")
		return nil, ee.ToError()
	}

	var checkData []ProbeData

	for _, r := range results {
		var cData ProbeData
		doc, err := bson.Marshal(r)
		if err != nil {
			ee.Message = "error marshaling results"
			ee.Error = err
			return nil, ee.ToError()
		}
		err = bson.Unmarshal(doc, &cData)
		if err != nil {
			ee.Message = "error unmarshalling results"
			ee.Error = err
			return nil, ee.ToError()
		}
		checkData = append(checkData, cData)
	}

	return checkData, nil
}

/*type MtrResult struct {
	StartTimestamp time.Time `json:"start_timestamp"bson:"start_timestamp"`
	StopTimestamp  time.Time `json:"stop_timestamp"bson:"stop_timestamp"`
	Triggered      bool      `bson:"triggered"json:"triggered"`
	Report         struct {
		Mtr struct {
			Src        string      `json:"src"bson:"src"`
			Dst        string      `json:"dst"bson:"dst"`
			Tos        interface{} `json:"tos"bson:"tos"`
			Tests      interface{} `json:"tests"bson:"tests"`
			Psize      string      `json:"psize"bson:"psize"`
			Bitpattern string      `json:"bitpattern"bson:"bitpattern"`
		} `json:"mtr"bson:"mtr"`
		Hubs []struct {
			Count interface{} `json:"count"bson:"count"`
			Host  string      `json:"host"bson:"host"`
			ASN   string      `json:"ASN"bson:"ASN"`
			Loss  float64     `json:"Loss%"bson:"Loss%"`
			Drop  int         `json:"Drop"bson:"Drop"`
			Rcv   int         `json:"Rcv"bson:"Rcv"`
			Snt   int         `json:"Snt"bson:"Snt"`
			Best  float64     `json:"Best"bson:"Best"`
			Avg   float64     `json:"Avg"bson:"Avg"`
			Wrst  float64     `json:"Wrst"bson:"Wrst"`
			StDev float64     `json:"StDev"bson:"StDev"`
			Gmean float64     `json:"Gmean"bson:"Gmean"`
			Jttr  float64     `json:"Jttr"bson:"Jttr"`
			Javg  float64     `json:"Javg"bson:"Javg"`
			Jmax  float64     `json:"Jmax"bson:"Jmax"`
			Jint  float64     `json:"Jint"bson:"Jint"`
		} `json:"hubs"bson:"hubs"`
	} `json:"report"bson:"report"`
}*/

type MtrResult struct {
	StartTimestamp time.Time `json:"start_timestamp" bson:"start_timestamp"`
	StopTimestamp  time.Time `json:"stop_timestamp" bson:"stop_timestamp"`
	Report         struct {
		Info struct {
			Target struct {
				IP       string `json:"ip" bson:"ip"`
				Hostname string `json:"hostname" bson:"hostname"`
			} `json:"target" bson:"target"`
		} `json:"info" bson:"info"`
		Hops []struct {
			TTL   int `json:"ttl" bson:"ttl"`
			Hosts []struct {
				IP       string `json:"ip" bson:"ip"`
				Hostname string `json:"hostname" bson:"hostname"`
			} `json:"hosts" bson:"hosts"`
			Extensions []string `json:"extensions" bson:"extensions"`
			LossPct    string   `json:"loss_pct" bson:"loss_pct"`
			Sent       int      `json:"sent" bson:"sent"`
			Last       string   `json:"last" bson:"last"`
			Recv       int      `json:"recv" bson:"recv"`
			Avg        string   `json:"avg" bson:"avg"`
			Best       string   `json:"best" bson:"best"`
			Worst      string   `json:"worst" bson:"worst"`
			StdDev     string   `json:"stddev" bson:"stddev"`
		} `json:"hops" bson:"hops"`
	} `json:"report" bson:"report"`
}

type NetResult struct {
	LocalAddress     string    `json:"local_address" bson:"local_address"`
	DefaultGateway   string    `json:"default_gateway" bson:"default_gateway"`
	PublicAddress    string    `json:"public_address" bson:"public_address"`
	InternetProvider string    `json:"internet_provider" bson:"internet_provider"`
	Lat              string    `json:"lat" bson:"lat"`
	Long             string    `json:"long" bson:"long"`
	Timestamp        time.Time `json:"timestamp" bson:"timestamp"`
}
type SpeedTestResult struct {
	TestData  []SpeedTestServer `json:"test_data"`
	Timestamp time.Time         `json:"timestamp" bson:"timestamp"`
}

type SpeedTestServer struct {
	URL          string                `json:"url,omitempty" bson:"url"`
	Lat          string                `json:"lat,omitempty" bson:"lat"`
	Lon          string                `json:"lon,omitempty" bson:"lon"`
	Name         string                `json:"name,omitempty" bson:"name"`
	Country      string                `json:"country,omitempty" bson:"country"`
	Sponsor      string                `json:"sponsor,omitempty" bson:"sponsor"`
	ID           string                `json:"id,omitempty" bson:"id"`
	Host         string                `json:"host,omitempty" bson:"host"`
	Distance     float64               `json:"distance,omitempty" bson:"distance"`
	Latency      time.Duration         `json:"latency,omitempty" bson:"latency"`
	MaxLatency   time.Duration         `json:"max_latency,omitempty" bson:"max_latency"`
	MinLatency   time.Duration         `json:"min_latency,omitempty" bson:"min_latency"`
	Jitter       time.Duration         `json:"jitter,omitempty" bson:"jitter"`
	DLSpeed      SpeedTestByteRate     `json:"dl_speed,omitempty" bson:"dl_speed"`
	ULSpeed      SpeedTestByteRate     `json:"ul_speed,omitempty" bson:"ul_speed"`
	TestDuration SpeedTestTestDuration `json:"test_duration,omitempty" bson:"test_duration"`
	PacketLoss   SpeedTestPLoss        `json:"packet_loss,omitempty" bson:"packet_loss"`
}

type SpeedTestByteRate float64

type SpeedTestTestDuration struct {
	Ping     *time.Duration `json:"ping" bson:"ping"`
	Download *time.Duration `json:"download" bson:"download"`
	Upload   *time.Duration `json:"upload" bson:"upload"`
	Total    *time.Duration `json:"total" bson:"total"`
}

type SpeedTestPLoss struct {
	Sent int `json:"sent" bson:"sent"` // Number of sent packets acknowledged by the remote.
	Dup  int `json:"dup" bson:"dup"`   // Number of duplicate packets acknowledged by the remote.
	Max  int `json:"max" bson:"max"`   // The maximum index value received by the remote.
}

type RPerfResults struct {
	StartTimestamp time.Time `json:"start_timestamp"bson:"start_timestamp"`
	StopTimestamp  time.Time `json:"stop_timestamp"bson:"stop_timestamp"`
	Config         struct {
		Additional struct {
			IpVersion   int  `json:"ip_version"bson:"ip_version"`
			OmitSeconds int  `json:"omit_seconds"bson:"omit_seconds"`
			Reverse     bool `json:"reverse"bson:"reverse"`
		} `json:"additional"bson:"additional"`
		Common struct {
			Family  string `json:"family"bson:"family"`
			Length  int    `json:"length"bson:"length"`
			Streams int    `json:"streams"bson:"streams"`
		} `json:"common"bson:"common"`
		Download struct {
		} `json:"download"bson:"download"`
		Upload struct {
			Bandwidth    int     `json:"bandwidth"bson:"bandwidth"`
			Duration     float64 `json:"duration"bson:"duration"`
			SendInterval float64 `json:"send_interval"bson:"send_interval"`
		} `json:"upload"bson:"upload"`
	} `json:"config"bson:"config"`
	/*Streams []struct {
		Abandoned bool `json:"abandoned"bson:"abandoned"`
		Failed    bool `json:"failed"bson:"failed"`
		Intervals struct {
			Receive []struct {
				BytesReceived     int     `json:"bytes_received"bson:"bytes_received"`
				Duration          float64 `json:"duration"bson:"duration"`
				JitterSeconds     float64 `json:"jitter_seconds"bson:"jitter_seconds"`
				PacketsDuplicated int     `json:"packets_duplicated"bson:"packets_duplicated"`
				PacketsLost       int     `json:"packets_lost"bson:"packets_lost"`
				PacketsOutOfOrder int     `json:"packets_out_of_order"bson:"packets_out_of_order"`
				PacketsReceived   int     `json:"packets_received"bson:"packets_received"`
				Timestamp         float64 `json:"timestamp"bson:"timestamp"`
				UnbrokenSequence  int     `json:"unbroken_sequence"bson:"unbroken_sequence"`
			} `json:"receive"bson:"receive"`
			Send []struct {
				BytesSent    int     `json:"bytes_sent"bson:"bytes_sent"`
				Duration     float64 `json:"duration"bson:"duration"`
				PacketsSent  int     `json:"packets_sent"bson:"packets_sent"`
				SendsBlocked int     `json:"sends_blocked"bson:"sends_blocked"`
				Timestamp    float64 `json:"timestamp"bson:"timestamp"`
			} `json:"send"bson:"send"`
			Summary struct {
				BytesReceived            int     `json:"bytes_received"bson:"bytes_received"`
				BytesSent                int     `json:"bytes_sent"bson:"bytes_sent"`
				DurationReceive          float64 `json:"duration_receive"bson:"duration_receive"`
				DurationSend             float64 `json:"duration_send"bson:"duration_send"`
				FramedPacketSize         int     `json:"framed_packet_size"bson:"framed_packet_size"`
				JitterAverage            float64 `json:"jitter_average"bson:"jitter_average"`
				JitterPacketsConsecutive int     `json:"jitter_packets_consecutive"bson:"jitter_packets_consecutive"`
				PacketsDuplicated        int     `json:"packets_duplicated"bson:"packets_duplicated"`
				PacketsLost              int     `json:"packets_lost"bson:"packets_lost"`
				PacketsOutOfOrder        int     `json:"packets_out_of_order"bson:"packets_out_of_order"`
				PacketsReceived          int     `json:"packets_received"bson:"packets_received"`
				PacketsSent              int     `json:"packets_sent"bson:"packets_sent"`
			} `json:"summary"bson:"summary"`
		} `json:"intervals"bson:"intervals"`
	} `json:"streams"bson:"streams"`*/
	Success bool `json:"success"bson:"success"`
	Summary struct {
		BytesReceived            int     `json:"bytes_received"bson:"bytes_received"`
		BytesSent                int     `json:"bytes_sent"bson:"bytes_sent"`
		DurationReceive          float64 `json:"duration_receive"bson:"duration_receive"`
		DurationSend             float64 `json:"duration_send"bson:"duration_send"`
		FramedPacketSize         int     `json:"framed_packet_size"bson:"framed_packet_size"`
		JitterAverage            float64 `json:"jitter_average"bson:"jitter_average"`
		JitterPacketsConsecutive int     `json:"jitter_packets_consecutive"bson:"jitter_packets_consecutive"`
		PacketsDuplicated        int     `json:"packets_duplicated"bson:"packets_duplicated"`
		PacketsLost              int     `json:"packets_lost"bson:"packets_lost"`
		PacketsOutOfOrder        int     `json:"packets_out_of_order"bson:"packets_out_of_order"`
		PacketsReceived          int     `json:"packets_received"bson:"packets_received"`
		PacketsSent              int     `json:"packets_sent"bson:"packets_sent"`
	} `json:"summary"bson:"summary"`
}

type PingResult struct {
	// StartTime is the time that the check started at
	StartTimestamp time.Time `json:"start_timestamp"bson:"start_timestamp"`
	StopTimestamp  time.Time `json:"stop_timestamp"bson:"stop_timestamp"`
	// PacketsRecv is the number of packets received.
	PacketsRecv int `json:"packets_recv"bson:"packets_recv"`
	// PacketsSent is the number of packets sent.
	PacketsSent int `json:"packets_sent"bson:"packets_sent"`
	// PacketsRecvDuplicates is the number of duplicate responses there were to a sent packet.
	PacketsRecvDuplicates int `json:"packets_recv_duplicates"bson:"packets_recv_duplicates"`
	// PacketLoss is the percentage of packets lost.
	PacketLoss float64 `json:"packet_loss"bson:"packet_loss"`
	// Addr is the string address of the host being pinged.
	Addr string `json:"addr"bson:"addr"`
	// MinRtt is the minimum round-trip time sent via this pinger.
	MinRtt time.Duration `json:"min_rtt"bson:"min_rtt"`
	// MaxRtt is the maximum round-trip time sent via this pinger.
	MaxRtt time.Duration `json:"max_rtt"bson:"max_rtt"`
	// AvgRtt is the average round-trip time sent via this pinger.
	AvgRtt time.Duration `json:"avg_rtt"bson:"avg_rtt"`
	// StdDevRtt is the standard deviation of the round-trip times sent via
	// this pinger.
	StdDevRtt time.Duration `json:"std_dev_rtt"bson:"std_dev_rtt"`
}

type CompleteSystemInfo struct {
	HostInfo   HostInfo       `json:"hostInfo" bson:"hostInfo"`
	MemoryInfo HostMemoryInfo `json:"memoryInfo" bson:"memoryInfo"`
	CPUTimes   CPUTimes       `json:"CPUTimes" bson:"CPUTimes"`
	Timestamp  time.Time      `json:"timestamp" bson:"timestamp"`
}

type CPUTimes struct {
	User    time.Duration `json:"user" bson:"user"`
	System  time.Duration `json:"system" bson:"system"`
	Idle    time.Duration `json:"idle,omitempty" bson:"idle"`
	IOWait  time.Duration `json:"iowait,omitempty" bson:"IOWait"`
	IRQ     time.Duration `json:"irq,omitempty" bson:"IRQ"`
	Nice    time.Duration `json:"nice,omitempty" bson:"nice"`
	SoftIRQ time.Duration `json:"soft_irq,omitempty" bson:"softIRQ"`
	Steal   time.Duration `json:"steal,omitempty" bson:"steal"`
}

type HostInfo struct {
	Architecture      string    `json:"architecture" bson:"architecture"`
	BootTime          time.Time `json:"boot_time" bson:"bootTime"`
	Containerized     *bool     `json:"containerized,omitempty" bson:"containerized"`
	Hostname          string    `json:"name" bson:"hostname"`
	IPs               []string  `json:"ip,omitempty" bson:"IPs"`
	KernelVersion     string    `json:"kernel_version" bson:"kernelVersion"`
	MACs              []string  `json:"mac" bson:"MACs"`
	OS                OSInfo    `json:"os" bson:"OS"`
	Timezone          string    `json:"timezone" bson:"timezone"`
	TimezoneOffsetSec int       `json:"timezone_offset_sec" bson:"timezoneOffsetSec"`
	UniqueID          string    `json:"id,omitempty" bson:"uniqueID"`
}

type OSInfo struct {
	Type     string `json:"type" bson:"type"`
	Family   string `json:"family" bson:"family"`
	Platform string `json:"platform" bson:"platform"`
	Name     string `json:"name" bson:"name"`
	Version  string `json:"version" bson:"version"`
	Major    int    `json:"major" bson:"major"`
	Minor    int    `json:"minor" bson:"minor"`
	Patch    int    `json:"patch" bson:"patch"`
	Build    string `json:"build,omitempty" bson:"build"`
	Codename string `json:"codename,omitempty" bson:"codename"`
}

// HostMemoryInfo (all values are specified in bytes).
type HostMemoryInfo struct {
	Total        uint64            `json:"total_bytes" bson:"total"`                // Total physical memory.
	Used         uint64            `json:"used_bytes" bson:"used"`                  // Total - Free
	Available    uint64            `json:"available_bytes" bson:"available"`        // Amount of memory available without swapping.
	Free         uint64            `json:"free_bytes" bson:"free"`                  // Amount of memory not used by the system.
	VirtualTotal uint64            `json:"virtual_total_bytes" bson:"virtualTotal"` // Total virtual memory.
	VirtualUsed  uint64            `json:"virtual_used_bytes" bson:"virtualUsed"`   // VirtualTotal - VirtualFree
	VirtualFree  uint64            `json:"virtual_free_bytes" bson:"virtualFree"`   // Virtual memory that is not used.
	Metrics      map[string]uint64 `json:"raw,omitempty" bson:"metrics"`            // Other memory related metrics.
}

/*type TrafficSimClientStats struct {
	AverageRTT       float64   `json:"averageRTT" bson:"averageRTT"`
	DuplicatePackets int       `json:"duplicatePackets" bson:"duplicatePackets"`
	LostPackets      int       `json:"lostPackets" bson:"lostPackets"`
	MaxRTT           int       `json:"maxRTT" bson:"maxRTT"`
	MinRTT           int       `json:"minRTT" bson:"minRTT"`
	OutOfSequence    int       `json:"outOfSequence" bson:"outOfSequence"`
	StdDevRTT        float64   `json:"stdDevRTT" bson:"stdDevRTT"`
	TotalPackets     int       `json:"totalPackets" bson:"totalPackets"`
	ReportTime       time.Time `json:"reportTime" bson:"reportTime"`
}*/

type TrafficSimClientStats struct {
	AverageRTT       float64 `json:"averageRTT" bson:"averageRTT"`
	DuplicatePackets int     `json:"duplicatePackets" bson:"duplicatePackets"`
	Flows            map[string]struct {
		BytesReceived int     `json:"bytesReceived"`
		BytesSent     int     `json:"bytesSent"`
		Direction     string  `json:"direction"`
		Duration      float64 `json:"duration"`
		JitterStats   struct {
			Min    int `json:"min"`
			Max    int `json:"max"`
			Avg    int `json:"avg"`
			StdDev int `json:"stdDev"`
		} `json:"jitterStats"`
		LossPercentage  int `json:"lossPercentage"`
		PacketsLost     int `json:"packetsLost"`
		PacketsReceived int `json:"packetsReceived"`
		PacketsSent     int `json:"packetsSent"`
		RttStats        struct {
			Min    int `json:"min"`
			Max    int `json:"max"`
			Avg    int `json:"avg"`
			StdDev int `json:"stdDev"`
			P50    int `json:"p50"`
			P95    int `json:"p95"`
			P99    int `json:"p99"`
		} `json:"rttStats"`
		ThroughputRecv float64 `json:"throughputRecv"`
		ThroughputSend float64 `json:"throughputSend"`
	} `json:"flows" bson:"flows"`
	LossPercentage int       `json:"lossPercentage" bson:"lossPercentage"`
	LostPackets    int       `json:"lostPackets" bson:"lostPackets"`
	MaxRTT         int       `json:"maxRTT" bson:"maxRTT"`
	MinRTT         int       `json:"minRTT" bson:"minRTT"`
	OutOfSequence  int       `json:"outOfSequence" bson:"outOfSequence"`
	ReportTime     time.Time `json:"reportTime" bson:"reportTime"`
	StdDevRTT      float64   `json:"stdDevRTT" bson:"stdDevRTT"`
	TotalPackets   int       `json:"totalPackets" bson:"totalPackets"`
}

/*return map[string]interface{}{
"lostPackets":      lostPackets,
"outOfSequence":    outOfOrder,
"duplicatePackets": duplicatePackets,
"averageRTT":       avgRTT,
"minRTT":           minRTT,
"maxRTT":           maxRTT,
"stdDevRTT":        stdDevRTT,
"totalPackets":     len(ts.ClientStats.PacketTimes),
}*/
