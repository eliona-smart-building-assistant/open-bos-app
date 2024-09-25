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

package broker

import (
	"fmt"
	appmodel "open-bos/app/model"
	"open-bos/eliona"
)

func GetDevices(config appmodel.Configuration) (eliona.Root, error) {
	client, err := NewOpenBOSClient(config.Gwid, config.ClientID, config.ClientSecret)
	if err != nil {
		return eliona.Root{}, fmt.Errorf("creating instance of client: %v", err)
	}
	_, err = client.getAssetTemplates()
	if err != nil {
		return eliona.Root{}, fmt.Errorf("getting functional block template: %v", err)
	}
	return eliona.Root{}, nil
}
