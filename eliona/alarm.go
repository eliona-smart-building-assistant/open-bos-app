//  This file is part of the Eliona project.
//  Copyright © 2025 IoTEC AG. All Rights Reserved.
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

package eliona

import (
	"fmt"
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v3"
	"github.com/eliona-smart-building-assistant/go-eliona/v2/client"
)

var CheckTypeExternal = "external"

func submitAlarm(apiKey string, alarmID *int32, assetID int32, subtype, attribute string, priority api.AlarmPriority, subject string, message map[string]any) (int32, error) {
	alarmRule, _, err := client.NewClient(client.ApiEndpointString()).AlarmRulesAPI.
		PutAlarmRule(client.AuthenticationContext(apiKey)).
		AlarmRule(api.AlarmRule{
			Id:        *api.NewNullableInt32(alarmID),
			AssetId:   assetID,
			Subtype:   api.DataSubtype(subtype),
			Attribute: attribute,
			Priority:  api.AlarmPriority(priority),
			Subject:   *api.NewNullableString(api.PtrString(subject)),
			Message:   message,
			Tags:      []string{},
			Enable:    api.PtrBool(true),
			CheckType: *api.NewNullableString(&CheckTypeExternal),
		}).
		Execute()
	if err != nil {
		return 0, err
	}
	return alarmRule.GetId(), nil
}

func CreateAlarm(apiKey string, assetID int32, subtype, attribute string, priority api.AlarmPriority, subject string, message map[string]any) (int32, error) {
	return submitAlarm(apiKey, nil, assetID, subtype, attribute, priority, subject, message)
}

func UpdateAlarm(apiKey string, alarmID int32, assetID int32, subtype, attribute string, priority api.AlarmPriority, subject string, message map[string]any) (int32, error) {
	return submitAlarm(apiKey, &alarmID, assetID, subtype, attribute, priority, subject, message)
}

func UpdateAlarmStatus(apiKey string, alarmID int32, message map[string]interface{}, appeared time.Time, ack bool, ackText string, closed bool) error {
	now := time.Now()

	alarm := api.Alarm{
		RuleId:    alarmID,
		Message:   message,
		Timestamp: *api.NewNullableTime(&appeared),
	}

	if ack {
		alarm.AcknowledgeTimestamp = *api.NewNullableTime(&now)
		alarm.AcknowledgeText = *api.NewNullableString(&ackText)
	}
	if closed {
		alarm.GoneTimestamp = *api.NewNullableTime(&now)
	}
	_, _, err := client.NewClient(client.ApiEndpointString()).AlarmsAPI.
		PutAlarm(client.AuthenticationContext(apiKey)).
		Alarm(alarm).
		Execute()
	if err != nil {
		return fmt.Errorf("updating alarm (%+v): %v", alarm, err)
	}
	return nil
}

func GetUserName(apiKey string, userID string) (string, error) {
	user, _, err := client.NewClient(client.ApiEndpointString()).UsersAPI.
		GetUserById(client.AuthenticationContext(apiKey), userID).
		Execute()
	if err != nil {
		return "", fmt.Errorf("getting user %v: %v", userID, err)
	}
	return fmt.Sprintf("%s %s", user.GetFirstname(), user.GetLastname()), nil
}
