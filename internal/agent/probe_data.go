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
	"time"
)

type ProbeData struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	ProbeID   primitive.ObjectID `json:"probe" bson:"probe"`
	Triggered bool               `json:"triggered" bson:"triggered"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
	Target    ProbeTarget        `bson:"target" json:"target"`
	Data      interface{}        `json:"data,omitempty" bson:"data,omitempty"`
}

func (pd *ProbeData) Create(db *mongo.Database) error {
	// todo handle to check if agent id is set and all that... or should it be in the api section??
	pd.ID = primitive.NewObjectID()

	pp := Probe{ID: pd.ProbeID}
	_, err := pp.Get(db)
	if err != nil {
		log.Error(err)
	}

	a := Agent{ID: pp.Agent}
	err = a.UpdateTimestamp(db)
	if err != nil {
		log.Error(err)
	}

	// load types

	pd.Data, _ = pd.parse(db)

	if (pd.CreatedAt == time.Time{}) {
		pd.CreatedAt = time.Now()
		log.Warnf("Timestamp was not included in probe data...")
	}

	mar, err := bson.Marshal(pd)
	if err != nil {
		log.Errorf("error marshalling check data when creating: %s", err)
		return err
	}

	var b *bson.D
	err = bson.Unmarshal(mar, &b)
	if err != nil {
		log.Errorf("error unmarhsalling check data when creating: %s", err)
		return err
	}
	result, err := db.Collection("probe_data").InsertOne(context.TODO(), b)
	if err != nil {
		log.Errorf("error inserting to database: %s", err)
		return err
	}

	fmt.Printf("created probe data with id: %v\n", result.InsertedID)
	return nil
}

func (pd *ProbeData) parse(db *mongo.Database) (interface{}, error) {
	// todo get ProbeType from ProbeID??
	pp := Probe{ID: pd.ProbeID}
	probe, err := pp.Get(db)
	if err != nil {
		return nil, err
	}

	switch probe[0].Type { // todo
	case ProbeType_RPERF:
		jsonData, err := json.Marshal(pd.Data)
		if err != nil {
			// Handle the error, perhaps return it
			return nil, err
		}

		var rperfData RPerfResults // Replace with the actual struct for RPERF data
		err = json.Unmarshal(jsonData, &rperfData)
		if err != nil {
			// Handle error
		}
		return rperfData, err

	case ProbeType_MTR:
		// First, marshal the interface{} back to JSON
		jsonData, err := json.Marshal(pd.Data)
		if err != nil {
			// Handle the error, perhaps return it
			return nil, err
		}

		// Now you can unmarshal the JSON into your struct
		var mtrData MtrResult
		err = json.Unmarshal(jsonData, &mtrData)
		if err != nil {
			// Handle the error, perhaps return it
			return nil, err
		}

		// Return the successfully unmarshaled data
		return mtrData, nil
	case ProbeType_NETWORKINFO:
		jsonData, err := json.Marshal(pd.Data)
		if err != nil {
			// Handle the error, perhaps return it
			return nil, err
		}

		var mtrData NetResult // Replace with the actual struct for MTR data
		err = json.Unmarshal(jsonData, &mtrData)
		if err != nil {
			// Handle error
		}
		return mtrData, err
	case ProbeType_PING:
		jsonData, err := json.Marshal(pd.Data)
		if err != nil {
			// Handle the error, perhaps return it
			return nil, err
		}

		var mtrData PingResult // Replace with the actual struct for MTR data
		err = json.Unmarshal(jsonData, &mtrData)
		if err != nil {
			// Handle error
		}
		return mtrData, err
	case ProbeType_SPEEDTEST:
		jsonData, err := json.Marshal(pd.Data)
		if err != nil {
			// Handle the error, perhaps return it
			return nil, err
		}

		var mtrData SpeedTestResult // Replace with the actual struct for MTR data
		err = json.Unmarshal(jsonData, &mtrData)
		if err != nil {
			// Handle error
		}
		return mtrData, err
	// Add cases for other probe types
	case ProbeType_SYSTEMINFO:
		jsonData, err := json.Marshal(pd.Data)
		if err != nil {
			// Handle the error, perhaps return it
			return nil, err
		}

		var mtrData CompleteSystemInfo // Replace with the actual struct for MTR data
		err = json.Unmarshal(jsonData, &mtrData)
		if err != nil {
			// Handle error
		}
		return mtrData, err

	default:
		// Handle unsupported probe types or return an error
	}

	return nil, nil
}

// GetData requires a checkrequest to be sent, if agent id is set,
// it will require the type to be sent in check, otherwise
// the check id will be used
func (c *Probe) GetData(req *ProbeDataRequest, db *mongo.Database) ([]*ProbeData, error) {
	opts := options.Find().SetLimit(req.Limit)

	// Combined filter
	var combinedFilter bson.M = bson.M{"probe": c.ID}
	if c.Agent != (primitive.ObjectID{0}) {
		combinedFilter["agent"] = c.Agent
		//combinedFilter["type"] = c.Type
	}

	if !req.Recent {
		opts.SetSort(bson.D{{"data.stop_timestamp", -1}})
		timeFilter := bson.M{
			"data.stop_timestamp": bson.M{
				"$gt": req.StartTimestamp,
				"$lt": req.EndTimestamp,
			},
		}
		for k, v := range timeFilter {
			combinedFilter[k] = v
		}
	} else {
		opts = opts.SetSort(bson.D{{"data.stop_timestamp", -1}})
	}

	cursor, err := db.Collection("probe_data").Find(context.TODO(), combinedFilter, opts)
	if err != nil {
		return nil, err
	}

	var results []bson.D
	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, errors.New("no data matches the provided check id")
	}

	var checkData []*ProbeData

	for _, r := range results {
		var cData ProbeData
		doc, err := bson.Marshal(r)
		if err != nil {
			log.Errorf("Error marshaling data: %s", err)
			return nil, err
		}
		err = bson.Unmarshal(doc, &cData)
		if err != nil {
			log.Errorf("Error unmarshaling data: %s", err)
			return nil, err
		}
		checkData = append(checkData, &cData)
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
	Triggered      bool      `json:"triggered" bson:"triggered"`
	Report         MtrReport `json:"report" bson:"report"`
}

type MtrReport struct {
	Mtr struct {
		Src        string `json:"src" bson:"src"`
		Dst        string `json:"dst" bson:"dst"`
		Tos        int    `json:"tos" bson:"tos"`     // Assuming 'tos' is an integer
		Tests      int    `json:"tests" bson:"tests"` // Assuming 'tests' is an integer
		Psize      string `json:"psize" bson:"psize"`
		Bitpattern string `json:"bitpattern" bson:"bitpattern"`
	} `json:"mtr" bson:"mtr"`
	Hubs []struct {
		Count int     `json:"count" bson:"count"` // Assuming 'count' is an integer
		Host  string  `json:"host" bson:"host"`
		ASN   string  `json:"ASN" bson:"ASN"`
		Loss  float64 `json:"Loss%" bson:"Loss%"`
		Drop  int     `json:"Drop" bson:"Drop"`
		Rcv   int     `json:"Rcv" bson:"Rcv"`
		Snt   int     `json:"Snt" bson:"Snt"`
		Best  float64 `json:"Best" bson:"Best"`
		Avg   float64 `json:"Avg" bson:"Avg"`
		Wrst  float64 `json:"Wrst" bson:"Wrst"`
		StDev float64 `json:"StDev" bson:"StDev"`
		Gmean float64 `json:"Gmean" bson:"Gmean"`
		Jttr  float64 `json:"Jttr" bson:"Jttr"`
		Javg  float64 `json:"Javg" bson:"Javg"`
		Jmax  float64 `json:"Jmax" bson:"Jmax"`
		Jint  float64 `json:"Jint" bson:"Jint"`
	} `json:"hubs" bson:"hubs"`
}
type NetResult struct {
	LocalAddress     string    `json:"local_address"bson:"local_address"`
	DefaultGateway   string    `json:"default_gateway"bson:"default_gateway"`
	PublicAddress    string    `json:"public_address"bson:"public_address"`
	InternetProvider string    `json:"internet_provider"bson:"internet_provider"`
	Lat              string    `json:"lat"bson:"lat"`
	Long             string    `json:"long"bson:"long"`
	Timestamp        time.Time `json:"timestamp"bson:"timestamp"`
}
type SpeedTestResult struct {
	Latency   time.Duration `json:"latency"bson:"latency"`
	DLSpeed   float64       `json:"dl_speed"bson:"dl_speed"`
	ULSpeed   float64       `json:"ul_speed"bson:"ul_speed"`
	Server    string        `json:"server"bson:"server"`
	Host      string        `json:"host"bson:"host"`
	Timestamp time.Time     `json:"timestamp"bson:"timestamp"`
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
	Streams []struct {
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
	} `json:"streams"bson:"streams"`
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
