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
	id                   bigserial primary key,
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
	configuration_id bigserial not null references weather_app.configuration(id) ON DELETE CASCADE,
	project_id       text      not null,
	global_asset_id  text      not null,
	provider_id      text      not null unique,
	is_root          boolean   not null default false,
	asset_id         integer
);

-- There is a transaction started in app.Init(). We need to commit to make the
-- new objects available for all other init steps.
-- Chain starts the same transaction again.
commit and chain;
