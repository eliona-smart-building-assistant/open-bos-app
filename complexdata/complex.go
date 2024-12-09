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

package complexdata

func DecodeComplexData(value map[string]any, parentPath string) map[string]any {
	flattened := make(map[string]any)
	for key, val := range value {
		currentPath := parentPath
		if currentPath != "" {
			currentPath += "." + key
		} else {
			currentPath = key
		}

		// Handle nested complex values recursively
		if nested, ok := val.(map[string]any); ok {
			nestedData := DecodeComplexData(nested, currentPath)
			for nestedKey, nestedVal := range nestedData {
				flattened[nestedKey] = nestedVal
			}
		} else {
			flattened[currentPath] = val
		}
	}
	return flattened
}
