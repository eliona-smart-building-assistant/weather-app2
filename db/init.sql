--  This file is part of the Eliona project.
--  Copyright © 2025 IoTEC AG. All Rights Reserved.
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

create schema if not exists weather_app;

-- Should be editable by eliona frontend.
create table if not exists weather_app.configuration
(
	id                   int primary key default 1 check (id = 1), -- only single configuration possible, due to assets not created by app
	api_key              text not null,
	refresh_interval     integer not null default 60,
	request_timeout      integer not null default 120,
	active               boolean not null default false,
	enable               boolean not null default false,
	project_ids          text[] not null,
	user_id              text not null
);

create table if not exists weather_app.asset
(
	id               bigserial primary key,
	project_id       text      not null,
	location_name    text      not null,
	lat              text      not null,
	lon              text      not null,
	asset_id         integer   not null unique
);

create table if not exists weather_app.root_asset
(
	id               int primary key,
	configuration_id int not null unique references weather_app.configuration(id) ON DELETE CASCADE,
	project_id       text      not null,
	gai              text      not null,
	asset_id         integer   not null unique
);

-- There is a transaction started in app.Init(). We need to commit to make the
-- new objects available for all other init steps.
-- Chain starts the same transaction again.
commit and chain;
