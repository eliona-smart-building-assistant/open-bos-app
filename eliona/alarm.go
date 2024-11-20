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

package eliona

import (
	"fmt"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/client"
)

var CHECK_TYPE_EXTERNAL = "external"

func CreateAlarm(assetID int32, subtype, attribute string, needsAck bool, priority int, message map[string]any) (int32, error) {
	alarmRule, _, err := client.NewClient().AlarmRulesAPI.
		PostAlarmRule(client.AuthenticationContext()).
		AlarmRule(api.AlarmRule{
			AssetId:             assetID,
			Subtype:             api.DataSubtype(subtype),
			Attribute:           attribute,
			Priority:            api.AlarmPriority(priority),
			RequiresAcknowledge: &needsAck,
			Message:             message,
			Tags:                []string{}, // todo
			Enable:              api.PtrBool(true),
			CheckType:           *api.NewNullableString(&CHECK_TYPE_EXTERNAL),
		}).
		Execute()
	if err != nil {
		return 0, fmt.Errorf("creating alarm: %v", err)
	}
	return alarmRule.GetId(), nil
}
