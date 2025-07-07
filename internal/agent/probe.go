package agent

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"nw-guardian/internal"
	"strings"
	"time"
)

/*

instead of making a separate module for agent probes, we can use this existing framework.
we use the existing one by:
	- Agent will be creating agent
	- Target[] is going to just be a list of agents that is stored in DB
	- There will be a separate sub menu for the agent probes that list all the target agents,
	with sub dashboards that also display the side by side comparison or similar with end to end traffic
	- Returning list of probes for agents running will get a little complicated, we only want to report
	traffic sim results in the opposite direction when running in server mode based on the traffic we collect
	- So it will report as a "reporting" ID in a sense for the probe data it self






*/

// probes are general single target probes
// agent probes are full completions of what is available (ping, mtr, and sim traffic)

type Probe struct {
	Type          ProbeType          `json:"type"bson:"type"`
	ID            primitive.ObjectID `json:"id"bson:"_id"`
	Agent         primitive.ObjectID `json:"agent"bson:"agent"`
	CreatedAt     time.Time          `bson:"createdAt"json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt"json:"updatedAt"`
	Notifications bool               `json:"notifications"bson:"notifications"` // notifications will be emailed to anyone who has permissions on their account / associated with the site
	Config        ProbeConfig        `bson:"config"json:"config"`
}

/*
when a list of probetargets is given, normal targets will only contain a target, and not an agent, etc
- this way we can then re-include the probetarget into the data it sends back to differentiate between targets
even though there is technically only 1 "probe"

*/

type ProbeConfig struct {
	Target   []ProbeTarget `json:"target" bson:"target"`
	Duration int           `json:"duration" bson:"duration"`
	Count    int           `json:"count" bson:"count"`
	Interval int           `json:"interval" bson:"interval"`
	Server   bool          `bson:"server" json:"server"`
	Pending  time.Time     `json:"pending" bson:"pending"` // timestamp of when it was made pending / invalidate it after 10 minutes or so?
}

// todo update targets to be a struct instead of a simple string

// ProbeTarget for group based target data, on  generation of the "targets" grabbed by the agent on connection
// it will grab the latest IPs of the agent and include those as the "target" it self to aide in automating
// ProbeTarget target string will automatically be populated if it is a group probe, if not, the normal target string will be used
type ProbeTarget struct {
	Target string             `json:"target,omitempty" bson:"target"`
	Agent  primitive.ObjectID `json:"agent,omitempty" bson:"agent"`
	Group  primitive.ObjectID `json:"group,omitempty" bson:"group"`
}

type ProbeAlert struct {
	Agent     primitive.ObjectID `json:"agent,omitempty" bson:"agent" bson:"agent"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
	Probe     Probe              `bson:"probe" json:"probe"`
	ProbeData ProbeData          `json:"probe_data" bson:"probeData"`
}

func DeleteProbesByAgentID(db *mongo.Database, agentID primitive.ObjectID) error {
	// todo if probe is deleted, delete associated data
	// todo if agent is delete, delete all probes, and data

	ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "probe.DeleteProbesByAgentID", ObjectID: agentID}

	p := Probe{Agent: agentID}
	get, err := p.Get(db)
	if err != nil {
		ee.Message = "unable to get probes by agent id"
		ee.Error = err
		return ee.ToError()
	}

	for _, probe := range get {
		err := DeleteProbeDataByProbeID(db, probe.ID)
		if err != nil {
			ee.Message = "error deleting probes by id"
			ee.Error = err
			return ee.ToError()
		}
	}

	// Convert the string ID to an ObjectID
	// Create a filter to match the document by ID
	filter := bson.M{"_id": agentID}

	// Perform the deletion
	_, err = db.Collection("probes").DeleteMany(context.TODO(), filter)
	if err != nil {
		ee.Message = "error deleting probes for agent"
		ee.Error = err
		return ee.ToError()
	}

	return nil
}

type ProbeType string

const (
	ProbeType_RPERF             ProbeType = "RPERF"
	ProbeType_MTR               ProbeType = "MTR"
	ProbeType_PING              ProbeType = "PING"
	ProbeType_SPEEDTEST         ProbeType = "SPEEDTEST"
	ProbeType_SPEEDTEST_SERVERS ProbeType = "SPEEDTEST_SERVERS"
	ProbeType_NETWORKINFO       ProbeType = "NETINFO"
	ProbeType_SYSTEMINFO        ProbeType = "SYSINFO"
	ProbeType_TRAFFICSIM        ProbeType = "TRAFFICSIM"
	ProbeType_AGENT             ProbeType = "AGENT" // this will be an array only used for internal calculations
)

type ProbeDataRequest struct {
	Limit          int64     `json:"limit"`
	StartTimestamp time.Time `json:"startTimestamp"`
	EndTimestamp   time.Time `json:"endTimestamp"`
	Recent         bool      `json:"recent"`
	Option         string    `json:"option"`
}

func (probe *Probe) FindSimilarProbes(db *mongo.Database) ([]*Probe, error) {
	ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "probe.FindSimilarProbes", ObjectID: probe.ID}

	if len(probe.Config.Target) == 0 {
		ee.Message = "no targets found in probe config"
		return nil, ee.ToError()
	}

	// Remove type before getting probes
	probe.Type = ""

	allProbes, err := probe.Get(db)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to fetch probes"
		return nil, ee.ToError()
	}

	similarProbes := findSimilarProbes(allProbes, probe)

	if len(similarProbes) == 0 {
		ee.Message = "no similar probes found"
		return nil, ee.ToError()
	}

	return similarProbes, nil
}

func findSimilarProbes(probes []*Probe, targetProbe *Probe) []*Probe {
	var similarProbes []*Probe

	for _, p := range probes {
		if len(p.Config.Target) == 0 {
			continue
		}

		for _, targetConfig := range targetProbe.Config.Target {
			if isSimilarProbe(p, targetConfig) {
				similarProbes = append(similarProbes, p)
				break // Move to the next probe once a match is found
			}
		}
	}

	return similarProbes
}

func isSimilarProbe(probe *Probe, targetConfig ProbeTarget) bool {
	for _, probeTarget := range probe.Config.Target {
		// Check for matching agent IDs
		if targetConfig.Agent != primitive.NilObjectID && targetConfig.Agent == probeTarget.Agent {
			return true
		}

		// Check for manual targets on the same agent
		if targetConfig.Agent == probeTarget.Agent &&
			targetConfig.Target != "" &&
			targetConfig.Target == probeTarget.Target {
			return true
		}

		// Check for matching group IDs (if implemented in the future)
		if targetConfig.Group != primitive.NilObjectID && targetConfig.Group == probeTarget.Group {
			return true
		}
	}
	return false
}

func (probe *Probe) Create(db *mongo.Database) error {
	ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "probe.Create", ObjectID: probe.ID}

	probe.ID = primitive.NewObjectID()
	probe.CreatedAt = time.Now()
	probe.UpdatedAt = time.Now()

	mar, err := bson.Marshal(probe)
	if err != nil {
		ee.Message = "unable to marshal probe"
		ee.Error = err
		return ee.ToError()
	}

	var b *bson.D
	err = bson.Unmarshal(mar, &b)
	if err != nil {
		ee.Message = "unable to unmarshal probe"
		ee.Error = err
		return ee.ToError()
	}
	_, err = db.Collection("probes").InsertOne(context.TODO(), b)
	if err != nil {
		ee.Message = "error inserting into probes"
		ee.Error = err
		return ee.ToError()
	}

	//fmt.Printf("created agent check with id: %v\n", result.InsertedID)
	return nil
}

func (probe *Probe) Get(db *mongo.Database) ([]*Probe, error) {
	ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "probe.Get"}

	var filter = bson.D{{"_id", probe.ID}}

	var objectID = probe.ID
	var objectType = "probe"

	if probe.Type != "" && probe.Agent != (primitive.ObjectID{0}) {
		filter = bson.D{{"agent", probe.Agent}, {"type", probe.Type}}
		objectID = probe.Agent
		objectType = "agent"
	} else if probe.Agent != (primitive.ObjectID{0}) {
		filter = bson.D{{"agent", probe.Agent}}
		objectID = probe.Agent
		objectType = "agent"
	}
	ee.ObjectID = objectID
	ee.Message = objectType + " - "

	cursor, err := db.Collection("probes").Find(context.TODO(), filter)
	if err != nil {
		ee.Error = err
		ee.Message += "unable to find probes"
		return nil, ee.ToError()
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		ee.Error = err
		ee.Message += "unable to cursor probes"
		return nil, ee.ToError()
	}

	//fmt.Println(results)
	var agentChecks []*Probe

	for _, r := range results {
		var acData Probe
		doc, err := bson.Marshal(r)
		if err != nil {
			ee.Error = err
			ee.Message += "error marshalling"
			return nil, ee.ToError()
		}
		err = bson.Unmarshal(doc, &acData)
		if err != nil {
			ee.Error = err
			ee.Message += "error unmarshalling"
			return nil, ee.ToError()
		}

		agentChecks = append(agentChecks, &acData)
	}

	return agentChecks, nil
}

// GetAll get all checks based on id, and &/or type
func (probe *Probe) GetAll(db *mongo.Database) ([]*Probe, error) {
	ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "probe.GetAll", ObjectID: probe.Agent}

	var filter = bson.D{{"agent", probe.Agent}}
	if probe.Type != "" {
		filter = bson.D{{"agent", probe.Agent}, {"type", probe.Type}}
	}

	cursor, err := db.Collection("probes").Find(context.TODO(), filter)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to find probes for agent"
		return nil, ee.ToError()
	}
	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		ee.Error = err
		ee.Message = "unable to cursor probes for agent"
		return nil, ee.ToError()
	}
	var agentCheck []*Probe

	for _, rb := range results {
		m, err := bson.Marshal(&rb)
		if err != nil {
			ee.Error = err
			ee.Message = "unable to marshal probes"
			return nil, ee.ToError()
		}
		var tC Probe
		err = bson.Unmarshal(m, &tC)
		if err != nil {
			ee.Error = err
			ee.Message = "unable to unmarshal probes for agent"
			return nil, ee.ToError()
		}
		agentCheck = append(agentCheck, &tC)
	}
	return agentCheck, nil
}

func (probe *Probe) GetAllProbesForAgent(db *mongo.Database) ([]*Probe, error) {
	ee := internal.ErrorFormat{
		Package:  "internal.agent",
		Level:    log.ErrorLevel,
		Function: "probe.GetAllProbesForAgent",
		ObjectID: probe.Agent,
	}

	// Build filter based on agent and optional type
	filter := bson.D{{"agent", probe.Agent}}
	if probe.Type != "" {
		filter = bson.D{{"agent", probe.Agent}, {"type", probe.Type}}
	}

	// Execute database query
	cursor, err := db.Collection("probes").Find(context.TODO(), filter)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to get probes for agent"
		return nil, ee.ToError()
	}

	var results []bson.D
	if err = cursor.All(context.TODO(), &results); err != nil {
		ee.Error = err
		ee.Message = "error retrieving cursor results"
		return nil, ee.ToError()
	}

	var agentProbes []*Probe
	for _, result := range results {
		probe, err := probe.unmarshalProbeResult(result, db, ee)
		if err != nil {
			log.Error(err)
			continue
		}
		agentProbes = append(agentProbes, probe)
	}

	return agentProbes, nil
}

// unmarshalProbeResult handles the unmarshaling and target resolution for a single probe result
func (p *Probe) unmarshalProbeResult(result bson.D, db *mongo.Database, ee internal.ErrorFormat) (*Probe, error) {
	// Marshal and unmarshal to convert bson.D to Probe struct
	marshaledData, err := bson.Marshal(&result)
	if err != nil {
		ee.Error = err
		ee.Message = "error marshalling probe result"
		return nil, ee.ToError()
	}

	var probe Probe
	if err = bson.Unmarshal(marshaledData, &probe); err != nil {
		ee.Error = err
		ee.Message = "error unmarshalling probe result"
		return nil, ee.ToError()
	}

	// Process probe based on its configuration and type
	if err := p.processProbeTargets(&probe, db); err != nil {
		return nil, err
	}

	return &probe, nil
}

// processProbeTargets handles target resolution for different probe types
func (p *Probe) processProbeTargets(probe *Probe, db *mongo.Database) error {
	// Handle traffic simulation server probes
	if probe.Config.Server && probe.Type == ProbeType_TRAFFICSIM {
		return p.processTrafficSimServer(probe, db)
	}

	// Handle probes with targets (excluding traffic sim servers)
	if len(probe.Config.Target) > 0 && !(probe.Config.Server && probe.Type == ProbeType_TRAFFICSIM) {
		return p.processProbeWithTargets(probe, db)
	}

	return nil
}

func (probe *Probe) copyProbe(probeType ProbeType, target ProbeTarget) (*Probe, error) {
	pCopy := *probe // dereference to copy the struct (value copy)
	pCopy.Type = probeType
	pCopy.Config.Target = []ProbeTarget{target}

	return &pCopy, nil // return the address of the copy
}

// generateFakeProbesForAgent creates virtual probes for agent-type probes
// These fake types will return the same format of probes, except the "Target" in the TargetGroups will be IP,
// and Agent will still be the original agent. The returned probedata by agent will replace the Target to contain
// the type of test along with the destination IP. The group in the probedata when returned will be the reporting agent for end to end analysis.
func (p *Probe) generateFakeProbesForAgent(probe *Probe, db *mongo.Database) error {
	var generatedProbes []*Probe

	for _, target := range probe.Config.Target {
		// Get target agent information
		agent := Agent{ID: target.Agent}
		if err := agent.Get(db); err != nil {
			log.Error("Failed to get agent:", err)
			continue
		}

		// Get target agent's public IP
		publicIP, err := p.getAgentPublicIP(agent.ID, agent, db)
		if err != nil {
			log.Error("Failed to get agent public IP:", err)
			continue
		}

		// Generate fake probes for each supported probe type
		probeTypes := []ProbeType{ProbeType_MTR, ProbeType_PING}

		// Check if traffic simulation is supported
		if p.isTrafficSimSupported(agent, db) {
			probeTypes = append(probeTypes, ProbeType_TRAFFICSIM)
		}

		for _, probeType := range probeTypes {
			fakeProbe, err := p.createFakeProbe(probe, probeType, target, publicIP, db)
			if err != nil {
				log.Error("Failed to create fake probe:", err)
				continue
			}

			if fakeProbe != nil {
				generatedProbes = append(generatedProbes, fakeProbe)
			}
		}
	}

	// Store generated probes for later retrieval (if needed)
	// This could be stored in memory, cache, or temporary collection
	// For now, we'll assume they're handled elsewhere in the pipeline

	return nil
}

// createFakeProbe creates a specific type of fake probe for an agent target
func (p *Probe) createFakeProbe(originalProbe *Probe, probeType ProbeType, target ProbeTarget, publicIP string, db *mongo.Database) (*Probe, error) {
	switch probeType {
	case ProbeType_MTR, ProbeType_PING:
		return p.createStandardFakeProbe(originalProbe, probeType, target, publicIP)
	case ProbeType_TRAFFICSIM:
		return p.createTrafficSimFakeProbe(originalProbe, target, publicIP, db)
	default:
		return nil, fmt.Errorf("unsupported probe type for fake probe generation: %s", probeType)
	}
}

// createStandardFakeProbe creates fake probes for MTR and PING types
func (p *Probe) createStandardFakeProbe(originalProbe *Probe, probeType ProbeType, target ProbeTarget, publicIP string) (*Probe, error) {
	fakeProbe, err := originalProbe.copyProbe(probeType, target)
	if err != nil {
		return nil, err
	}

	// Set the target IP address
	fakeProbe.Config.Target[0].Target = publicIP

	// Preserve the original agent ID for tracking
	fakeProbe.Config.Target[0].Agent = target.Agent

	return fakeProbe, nil
}

// createTrafficSimFakeProbe creates fake probes for traffic simulation
func (p *Probe) createTrafficSimFakeProbe(originalProbe *Probe, target ProbeTarget, publicIP string, db *mongo.Database) (*Probe, error) {
	fakeProbe, err := originalProbe.copyProbe(ProbeType_TRAFFICSIM, target)
	if err != nil {
		return nil, err
	}

	// Find the traffic simulation server on the target agent
	serverProbe := Probe{Agent: target.Agent, Type: ProbeType_TRAFFICSIM}
	agentProbes, err := serverProbe.GetAll(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get traffic sim probes for agent %s: %w", target.Agent.Hex(), err)
	}

	// Find the server instance and configure target
	for _, agentProbe := range agentProbes {
		if agentProbe.Config.Server && agentProbe.Type == ProbeType_TRAFFICSIM {
			// Extract port from server target configuration
			targetParts := strings.Split(agentProbe.Config.Target[0].Target, ":")
			if len(targetParts) < 2 {
				continue // Skip invalid configurations
			}

			port := targetParts[1]
			fakeProbe.Config.Target[0].Target = publicIP + ":" + port
			fakeProbe.Config.Target[0].Agent = target.Agent

			return fakeProbe, nil
		}
	}

	// No traffic sim server found, return nil (not an error, just not supported)
	return nil, nil
}

// isTrafficSimSupported checks if the target agent supports traffic simulation
func (p *Probe) isTrafficSimSupported(agent Agent, db *mongo.Database) bool {
	// Check if the agent has any traffic simulation server probes
	serverProbe := Probe{Agent: agent.ID, Type: ProbeType_TRAFFICSIM}
	agentProbes, err := serverProbe.GetAll(db)
	if err != nil {
		return false
	}

	// Look for at least one server instance
	for _, probe := range agentProbes {
		if probe.Config.Server && probe.Type == ProbeType_TRAFFICSIM {
			return true
		}
	}

	return false
}
func (p *Probe) processTrafficSimServer(probe *Probe, db *mongo.Database) error {
	clients, err := FindTrafficSimClients(db, probe.Agent)
	if err != nil {
		return err
	}

	// Add client agents as targets (preserving original binding target)
	for _, client := range clients {
		newTarget := ProbeTarget{Agent: client.Agent}
		probe.Config.Target = append(probe.Config.Target, newTarget)
	}

	return nil
}

// processProbeWithTargets handles probes that have target configurations
func (p *Probe) processProbeWithTargets(probe *Probe, db *mongo.Database) error {
	// Handle agent-type probes
	if probe.Type == ProbeType_AGENT {
		return p.generateFakeProbesForAgent(probe, db)
	}

	// Handle probes targeting other agents
	if probe.Config.Target[0].Agent != (primitive.ObjectID{}) {
		return p.resolveAgentTarget(probe, db)
	}

	return nil
}

// resolveAgentTarget resolves the target IP address for agent-targeted probes
func (p *Probe) resolveAgentTarget(probe *Probe, db *mongo.Database) error {
	targetAgent := Agent{ID: probe.Config.Target[0].Agent}
	if err := targetAgent.Get(db); err != nil {
		return err
	}

	// Get target agent's public IP
	publicIP, err := p.getAgentPublicIP(probe.Config.Target[0].Agent, targetAgent, db)
	if err != nil {
		return err
	}

	// Handle different probe types
	switch probe.Type {
	case ProbeType_RPERF, ProbeType_TRAFFICSIM:
		return p.configureRPerfTarget(probe, publicIP, db)
	default:
		probe.Config.Target[0].Target = publicIP
	}

	return nil
}

// getAgentPublicIP retrieves the public IP address for a target agent
func (p *Probe) getAgentPublicIP(agentID primitive.ObjectID, agent Agent, db *mongo.Database) (string, error) {
	// Use override if available
	if agent.PublicIPOverride != "" {
		return agent.PublicIPOverride, nil
	}

	// Get network info from agent
	networkProbe := Probe{Agent: agentID, Type: ProbeType_NETWORKINFO}
	probeData, err := networkProbe.Get(db)
	if err != nil {
		return "", err
	}

	if len(probeData) == 0 {
		return "", fmt.Errorf("no network data found for agent %s", agentID.Hex())
	}

	// Get most recent data
	probeData[0].Agent = primitive.ObjectID{}
	data, err := probeData[0].GetData(&ProbeDataRequest{Recent: true, Limit: 1}, db)
	if err != nil {
		return "", err
	}

	if len(data) == 0 {
		return "", fmt.Errorf("no recent network data found for agent %s", agentID.Hex())
	}

	// Extract public IP from the most recent data
	lastElement := data[len(data)-1]
	netResult, err := p.extractNetResult(lastElement.Data)
	if err != nil {
		return "", err
	}

	return netResult.PublicAddress, nil
}

// extractNetResult extracts NetResult from probe data
func (p *Probe) extractNetResult(data interface{}) (NetResult, error) {
	var netResult NetResult

	switch v := data.(type) {
	case primitive.D:
		bsonData, err := bson.Marshal(v)
		if err != nil {
			return netResult, err
		}
		err = bson.Unmarshal(bsonData, &netResult)
		return netResult, err

	case primitive.M:
		bsonData, err := bson.Marshal(v)
		if err != nil {
			return netResult, err
		}
		err = bson.Unmarshal(bsonData, &netResult)
		return netResult, err

	default:
		return netResult, fmt.Errorf("data is neither primitive.D nor primitive.M")
	}
}

// configureRPerfTarget configures the target for RPerf and TrafficSim probes
func (p *Probe) configureRPerfTarget(probe *Probe, publicIP string, db *mongo.Database) error {
	// Find the corresponding server probe for this agent
	serverProbe := Probe{Agent: probe.Config.Target[0].Agent, Type: probe.Type}
	agentProbes, err := serverProbe.GetAll(db)
	if err != nil {
		return err
	}

	// Find the server instance and extract port
	for _, agentProbe := range agentProbes {
		if agentProbe.Config.Server && agentProbe.Type == probe.Type {
			// Extract port from server target configuration
			targetParts := strings.Split(agentProbe.Config.Target[0].Target, ":")
			if len(targetParts) < 2 {
				return fmt.Errorf("invalid server target format: %s", agentProbe.Config.Target[0].Target)
			}

			port := targetParts[1]
			probe.Config.Target[0].Target = publicIP + ":" + port
			return nil
		}
	}

	return fmt.Errorf("no server probe found for agent %s with type %s",
		probe.Config.Target[0].Agent.Hex(), probe.Type)
}

func (probe *Probe) UpdateFirstProbeTarget(db *mongo.Database, targetStatus string) error {
	ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "probe.UpdateFirstProbeTarget", ObjectID: probe.Agent}
	var filter = bson.D{{"_id", probe.ID}}

	get, err := probe.Get(db)
	if err != nil {
		return err
	}
	get[0].Config.Target[0].Target = targetStatus

	if get[0].Type == ProbeType_SPEEDTEST {
		get[0].Config.Pending = time.Now()
	}

	update := bson.D{
		{"$set", get[0]},
	}

	_, err = db.Collection("probes").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ee.Error = err
		ee.Message = "failed to update doc"
	}

	return nil
}

func FindTrafficSimClients(db *mongo.Database, serverAgentID primitive.ObjectID) ([]*Probe, error) {
	ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "probe.FindTrafficSimClients", ObjectID: serverAgentID}

	// we shouldn't need to search based on all the sites / reduce them because we trust the backend / no one will
	// abuse the functionality of adding a traffic sim server to a site that it doesn't belong to / link to agent?

	// Assuming `serverAgentID` is the ID of the agent with the TRAFFICSIM server probe.

	// Step 1: Define the filter to find non-server TRAFFICSIM probes targeting this server agent.
	filter := bson.D{
		{"type", ProbeType_TRAFFICSIM},         // Filter for TRAFFICSIM type probes.
		{"config.server", false},               // Ensure these are not servers.
		{"config.target.agent", serverAgentID}, // Target must be the server agent.
	}

	// Step 2: Query the probes collection based on the defined filter.
	var clientProbes []*Probe
	cursor, err := db.Collection("probes").Find(context.TODO(), filter)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to get traffic sim clients 1"
		return nil, ee.ToError()
	}
	if err := cursor.All(context.TODO(), &clientProbes); err != nil {
		ee.Error = err
		ee.Message = "unable to get traffic sim clients 2"
		return nil, ee.ToError()
	}

	// `clientProbes` now contains all TRAFFICSIM client probes targeting the given server agent.
	return clientProbes, nil
}

func (probe *Probe) Update(db *mongo.Database) error {
	ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "probe.Update", ObjectID: probe.ID}

	var filter = bson.D{{"_id", probe.ID}}

	marshal, err := bson.Marshal(probe)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to marshal"
		return ee.ToError()
	}

	var b bson.D
	err = bson.Unmarshal(marshal, &b)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to unmarshal"
		return ee.ToError()
	}

	update := bson.D{{"$set", b}}

	_, err = db.Collection("probes").UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to update probe"
		return ee.ToError()
	}

	return nil
}

// Delete check based on provided agent ID in check struct
func (probe *Probe) Delete(db *mongo.Database) error {
	ee := internal.ErrorFormat{Package: "internal.agent", Level: log.ErrorLevel, Function: "probe.Delete", ObjectID: probe.ID}
	// filter based on check ID
	var filter = bson.D{{"_id", probe.ID}}
	if (probe.Agent != primitive.ObjectID{}) {
		filter = bson.D{{"agent", probe.Agent}}
	}

	_, err := db.Collection("probes").DeleteMany(context.TODO(), filter)
	if err != nil {
		ee.Error = err
		ee.Message = "unable to delete probe"
		return ee.ToError()
	}

	return nil
}
