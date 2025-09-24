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

package appmodel

type Configuration struct {
	ElionaTenantId  string
	ElionaSiteId    string
	Id              int64
	GwId            string
	ClientId        string
	ClientSecret    string
	OntologyVersion int32
	AppPublicApiUrl string
	RefreshInterval int32
	RequestTimeout  int32
	AssetFilter     [][]FilterRule
	Enable          bool
	Active          bool
	UserId          *string
	ApiKey          string
}

type FilterRule struct {
	Parameter string
	Regex     string
}

type Asset struct {
	ID            int64
	Config        Configuration
	GlobalAssetID string
	ProviderID    string
	AssetID       int32
}

type Datapoint struct {
	ProviderID          string
	Subtype             string
	Asset               *Asset
	AttributeNamePrefix string
	Attributes          []Attribute

	Data map[string]any // For passing data of properties during ontology sync.
}

type Attribute struct {
	ID   int64
	Name string
}

type Alarm struct {
	ElionaAlarmID     int32
	ElionaAttributeID int64
	OpenBOSAlarmID    string
}
