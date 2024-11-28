//  This file is part of the Eliona project.
//  Copyright © 2024 IoTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package webhook

import (
	"open-bos/app"

	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/log"
)

type webhookServer struct {
	mux *http.ServeMux
}

func newWebhookServer() *webhookServer {
	return &webhookServer{
		mux: http.NewServeMux(),
	}
}

func (s *webhookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug("webhook", "Received request for URL: %s, Method: %s", r.URL.Path, r.Method)

	configID, err := parseConfigIDFromPath(r.URL.Path)
	if err != nil {
		log.Warn("webhook", "Invalid URL path, missing or invalid config ID: %s", r.URL.Path)
		http.Error(w, "Invalid URL path, missing or invalid config ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, "configID", configID)
	r = r.WithContext(ctx)

	r.URL.Path = removeConfigIDFromPath(r.URL.Path)

	s.mux.ServeHTTP(w, r)
}

func (s *webhookServer) handleOntologyVersion(w http.ResponseWriter, r *http.Request) {
	configID := r.Context().Value("configID").(int64)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	log.Debug("webhook", "Received ontology request headers: %+v", r.Header)
	log.Debug("webhook", "Request body: %s", body)
	log.Debug("webhook", "Method: %s", r.Method)
	log.Debug("webhook", "Config ID: %d", configID)

	// TODO: Implement version parsing once we know the format of the data.

	app.CollectConfigData(configID)
}

func (s *webhookServer) handleLivedataUpdate(w http.ResponseWriter, r *http.Request) {
	configID := r.Context().Value("configID").(int64)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	log.Debug("webhook", "Received data request headers: %+v", r.Header)
	log.Debug("webhook", "Request body: %s", body)
	log.Debug("webhook", "Method: %s", r.Method)
	log.Debug("webhook", "Config ID: %d", configID)

	type LiveData struct {
		DatapointID string `json:"Id"`
		IsProperty  bool   `json:"IsProperty"`
		TimeStamp   string `json:"TimeStamp"`
		Quality     string `json:"Quality"`
		Value       any    `json:"Value"`
	}

	var liveData []LiveData
	if err := json.Unmarshal(body, &liveData); err != nil {
		log.Error("webhook", "Failed to parse request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, data := range liveData {
		// Timestamps are always in UTC - see docs
		layout := "02/01/2006 15:04:05"
		timestamp, err := time.ParseInLocation(layout, data.TimeStamp, time.UTC)
		if err != nil {
			log.Warn("webhook", "Invalid timestamp format for ID %s: %v", data.DatapointID, err)
			continue
		}

		if data.Quality == "good" {
			app.UpdateDataPointInEliona(app.AttributeDataUpdate{
				ConfigID:            configID,
				DatapointProviderID: data.DatapointID,
				Timestamp:           timestamp,
				Value:               data.Value,
			})
		} else {
			log.Debug("webhook", "Received bad quality data for ID %s: IsProperty=%v, TimeStamp=%v, Quality=%s, Value=%v",
				data.DatapointID, data.IsProperty, timestamp, data.Quality, data.Value)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (s *webhookServer) handleLiveAlarm(w http.ResponseWriter, r *http.Request) {
	configID := r.Context().Value("configID").(int64)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	log.Debug("webhook", "Received alarm request headers: %+v", r.Header)
	log.Debug("webhook", "Request body: %s", body)
	log.Debug("webhook", "Method: %s", r.Method)
	log.Debug("webhook", "Config ID: %d", configID)

	type LiveAlarm struct {
		DataPointInstanceId string   `json:"dataPointInstanceId"` // DataPointInstanceId: Id of datapoint that caused the alarm. Nullable.
		SessionId           string   `json:"sessionId"`           // SessionId: Id of the alarm. Called sessionId and not id because for a single alarm you can receive several events. Nullable.
		Name                string   `json:"name"`                // Name: Name of the alarm. Nullable.
		Description         string   `json:"description"`         // Description: Description of the alarm. Nullable.
		Trigger             string   `json:"trigger"`             // Trigger: Trigger type of the alarm. Can be: Analognotvalue, analogvalue, digitaloff, digitalon, analogoutband2, analogoutband1, analoginband2, analoginband1, analoglo, analoglolo, analoghi, analoghihi, networkerror. Nullable.
		Active              bool     `json:"active"`              // Active: True if still active on the bus.
		Acked               bool     `json:"acked"`               // Acked: True if already acked.
		Closed              bool     `json:"closed"`              // Closed: True if alarm is closed. This is true ONLY for an event during a subscription to notify the alarm disappears.
		TimeStamp           string   `json:"timeStamp"`           // TimeStamp: UTC timestamp of the apparition of the alarm. Nullable.
		Quality             string   `json:"quality"`             // Quality: Quality of the value that caused the alarm. "Good" for a valid value, "bad..." for a bad quality. Nullable.
		Value               any      `json:"value"`               // Value: Value that caused the alarm. Value format depends on the DataType of the datapoint instance. Nullable.
		AckedBy             string   `json:"ackedBy"`             // AckedBy: The user who acknowledged the alarm. Nullable.
		Comment             string   `json:"comment"`             // Comment: Comment added when acknowledging the alarm. Nullable.
		NeedAcknowledge     bool     `json:"needAcknowledge"`     // NeedAcknowledge: True if alarm requires an ack.
		Severity            string   `json:"severity"`            // Severity: Severity of the alarm. Can be: Log, Low, High, Urgent, Critical. Nullable.
		AssetId             string   `json:"assetId"`             // AssetId: Id of the asset the alarm is attached to. Relevant especially for alarm attached to an orphan datapoint. Nullable.
		SpaceId             string   `json:"spaceId"`             // SpaceId: Id of the space the alarm is attached to. Relevant especially for alarm attached to an orphan datapoint. Nullable.
		AssetName           string   `json:"assetName"`           // AssetName: Name of the asset the alarm is attached to. Only if datapoint belongs to an asset. Nullable.
		SpaceName           string   `json:"spaceName"`           // SpaceName: Name of the space the alarm is attached to. Only if datapoint belongs to a space. Nullable.
		DatapointName       string   `json:"datapointName"`       // DatapointName: Name of the datapoint the alarm is attached to. Nullable.
		UnitSymbol          string   `json:"unitSymbol"`          // UnitSymbol: Unit symbol of the value. Nullable.
		Tags                []string `json:"tags"`                // Tags: Tags of the datapoint plus space/asset. Nullable.
	}

	var liveAlarms []LiveAlarm
	if err := json.Unmarshal(body, &liveAlarms); err != nil {
		log.Error("webhook", "Failed to parse request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, alarm := range liveAlarms {
		// Timestamps are always in UTC - see docs
		layout := "02/01/2006 15:04:05"
		timestamp, err := time.ParseInLocation(layout, alarm.TimeStamp, time.UTC)
		if err != nil {
			log.Warn("webhook", "Invalid timestamp format for alarm SessionId %s: %v", alarm.SessionId, err)
			continue
		}

		if alarm.Quality != "good" {
			log.Debug("webhook", "Received alarm with bad quality for SessionId %s: Quality=%s", alarm.SessionId, alarm.Quality)
			continue
		}

		alarmUpdate := app.AlarmUpdate{
			ConfigID:            configID,
			AlarmID:             alarm.SessionId,
			DatapointInstanceId: alarm.DataPointInstanceId,
			Timestamp:           timestamp,
			Severity:            alarm.Severity,
			Active:              alarm.Active,
			Acked:               alarm.Acked,
			Closed:              alarm.Closed,
			Name:                alarm.Name,
			Description:         alarm.Description,
			Value:               alarm.Value,
			AckedBy:             alarm.AckedBy,
			Comment:             alarm.Comment,
			NeedAcknowledge:     alarm.NeedAcknowledge,
			AssetId:             alarm.AssetId,
			SpaceId:             alarm.SpaceId,
			AssetName:           alarm.AssetName,
			SpaceName:           alarm.SpaceName,
			DatapointName:       alarm.DatapointName,
			UnitSymbol:          alarm.UnitSymbol,
			Tags:                alarm.Tags,
		}

		app.UpdateAlarmInEliona(alarmUpdate)
	}

	w.WriteHeader(http.StatusOK)
}

func parseConfigIDFromPath(path string) (int64, error) {
	// Matches "/{configID}/rest-of-path"
	re := regexp.MustCompile(`^/(\d+)/`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 2 {
		return 0, fmt.Errorf("config ID not found in path")
	}
	return strconv.ParseInt(matches[1], 10, 64)
}

func removeConfigIDFromPath(path string) string {
	re := regexp.MustCompile(`^/\d+/`)
	return re.ReplaceAllString(path, "/")
}

func StartWebhookListener() {
	server := newWebhookServer()

	server.mux.HandleFunc("/ontology-version", server.handleOntologyVersion)
	server.mux.HandleFunc("/ontology-livedata", server.handleLivedataUpdate)
	server.mux.HandleFunc("/ontology-livealarm", server.handleLiveAlarm)

	http.Handle("/", server)
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal("webhook", "Error starting server on port 8081: %v\n", err)
	}
}
