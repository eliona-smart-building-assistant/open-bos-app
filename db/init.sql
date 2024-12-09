--  This file is part of the Eliona project.
--  Copyright © 2024 IoTEC AG. All Rights Reserved.
--  ______ _ _
-- |  ____| (_)
-- | |__  | |_  ___  _ __   __ _
-- |  __| | | |/ _ \| '_ \ / _` |
-- | |____| | | (_) | | | | (_| |
-- |______|_|_|\___/|_| |_|\__,_|
--
--  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
--  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
--  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
--  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
--  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

create schema if not exists open_bos;

-- Should be editable by eliona frontend.
create table if not exists open_bos.configuration
(
	id                   bigserial primary key,
	gwid                 text not null,
	client_id            text not null,
	client_secret        text not null,
	ontology_version     integer not null,
	app_public_api_url   text not null,
	refresh_interval     integer not null default 60,
	request_timeout      integer not null default 120,
	asset_filter         json not null,
	active               boolean not null default false,
	enable               boolean not null default false,
	project_ids          text[] not null,
	user_id              text not null
);

create table if not exists open_bos.asset
(
	id               bigserial primary key,
	configuration_id bigserial not null references open_bos.configuration(id) ON DELETE CASCADE,
	project_id       text      not null,
	global_asset_id  text      not null,
	provider_id      text      not null,
	asset_id         integer
);

create table if not exists open_bos.openbos_datapoint
(
	id          bigserial primary key,
	asset_id    bigserial not null references open_bos.asset(id) ON DELETE CASCADE,
	subtype     text      not null,
	provider_id text      not null unique,
	name        text      not null
);

create table if not exists open_bos.eliona_attribute
(
	id                    bigserial primary key,
	openbos_datapoint_id  bigserial not null references open_bos.openbos_datapoint(id) ON DELETE CASCADE,
	eliona_attribute_name text      not null
);

CREATE TABLE IF NOT EXISTS open_bos.alarm (
	id                    BIGSERIAL PRIMARY KEY,
	eliona_attribute_id   BIGSERIAL NOT NULL REFERENCES open_bos.eliona_attribute(id) ON DELETE CASCADE,
	eliona_alarm_id       INTEGER NOT NULL unique,
	openbos_alarm_id      TEXT NOT NULL -- Not unique, because of complex mapping to multiple attributes.
);

-- There is a transaction started in app.Init(). We need to commit to make the
-- new objects available for all other init steps.
-- Chain starts the same transaction again.
commit and chain;
