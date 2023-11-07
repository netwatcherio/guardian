package workers

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"nw-guardian/internal/agent"
)

/*
func Push2Loki(cd *agent.Data, db *mongo.Database) error {
	// List of default labels which would be attached to every log message
	agent := agent.Agent{ID: cd.AgentID}
	err := agent.Get(db)
	if err != nil {
		return err
	}

	if agent.LokiDataPath == "" {
		return errors.New("agent does not have loki data path set")
	}

	site2 := site.Site{ID: agent.Site}
	err = site2.Get(db)
	if err != nil {
		return err
	}

	identifiers := map[string]string{
		"agentName":   agent.Name,
		"agentID":     agent.ID.Hex(),
		"siteID":      agent.Site.Hex(),
		"siteName":    site2.Name,
		"checkType":   string(cd.Type),
		"checkTarget": cd.Target,
	}

	promtailClient, err := promtail.NewJSONv1Client(agent.LokiDataPath, identifiers)
	if err != nil {
		return err
	}
	defer promtailClient.Close()

	marshal, err := json.Marshal(cd)
	if err != nil {
		return err
	}
	str := strings.Replace(string(marshal), "%", "", -1)

	promtailClient.LogfWithLabels(promtail.Info, identifiers, str)

	return nil

	/*customLabels := map[string]string{
		"somethingSpecial": "right-here",
	}
		promtailClient.LogfWithLabels(promtail.Info, customLabels, "Still here")*/
//}*/

func CreateProbeDataWorker(c chan agent.ProbeData, db *mongo.Database) {
	go func(cc chan agent.ProbeData) {
		for {
			data := <-cc

			err := data.Create(db)
			if err != nil {
				log.Error(err)
			}
		}
	}(c)
}

/*func CreateCheckWorker(c chan agent.Data, db *mongo.Database) {
	go func(cl chan agent.Data) {
		log.Info("Starting check data creation worker...")
		for {
			data := <-cl

			// TODO send data directly to loki data path, and also add it to it's own DB :)

			err := data.Create(db)
			if err != nil {
				log.Error(err)
			}

			if data.Type == agent.CtSpeedTest {
				agentC := agent.Check{ID: data.CheckID}
				_, err := agentC.Get(db)
				if err != nil {
					log.Error(err)
					return
				}
				agentC.Pending = false
				err = agentC.Update(db)
				if err != nil {
					log.Error(err)
					return
				}
			}

			err = Push2Loki(&data, db)
			if err != nil {
				log.Error(err)
			}
		}
	}(c)
}*/
