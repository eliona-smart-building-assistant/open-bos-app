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

package main

import (
	"context"
	elionabackend "github.com/eliona-smart-building-assistant/backend-frm/pkg/eliona"
	"github.com/eliona-smart-building-assistant/backend-frm/pkg/postgres"
	elionaapp "github.com/eliona-smart-building-assistant/go-eliona/v2/app"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"open-bos/v2/app"
	dbhelper "open-bos/v2/db/helper"
	"open-bos/v2/webhook"
	"time"
)

// The main function starts the app by starting all services necessary for this app and waits
// until all services are finished.
func main() {
	log.Info("main", "Starting the app.")

	// Set default database to use boil.*G functions.
	pool, err := elionabackend.GetDatabasePoolWithOverrideRole(context.Background(), "eliona", "api-v2", 10, postgres.WithResetOnAcquire())
	if err != nil {
		log.Fatal("Database", "Cannot open database: %v", err)
	}
	defer pool.Close(context.Background())
	database := pool.StdlibDB()
	boil.SetDB(database)

	// Set the database logging level.
	if log.Lev() >= log.TraceLevel {
		boil.DebugMode = true
		boil.DebugWriter = log.GetWriter(log.TraceLevel, "database")
	}

	// TODO: App init inside apps is no longer supported. MUST be replaced by new multi tenancy app concept.
	// Initialize the app
	// app.Initialize()

	// Fetch the API keys configured for the app
	dbhelper.FetchApiKeys(elionaapp.AppName(), database)

	// Starting the service to collect the data for this app.
	common.WaitForWithOs(
		common.Loop(app.CollectData, time.Second),
		app.ListenApi,
		app.ListenForOutputChanges,
		app.ListenForAlarmChanges,
		webhook.StartWebhookListener,
		common.Loop(app.Heartbeat, 2*time.Minute),
	)

	log.Info("main", "Cancelling subscriptions...")
	app.Teardown()
	log.Info("main", "Terminating the app.")
}
