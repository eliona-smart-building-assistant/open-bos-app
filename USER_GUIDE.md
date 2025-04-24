# OpenBOS User Guide

### Introduction

> The OpenBOS app provides integration and synchronization between Eliona and [ABB Ability™ Building Edge](https://new.abb.com/low-voltage/products/building-automation/building-edge) (using OpenBOS software).

## Overview

This guide provides instructions on configuring, installing, and using the OpenBOS app to manage resources and synchronize data between Eliona and OpenBOS edge.

## Steps needed to be taken in OpenBOS Building Edge

The complete documentation for using Edge Editor, embedded tool of the Building Edge allowing its configuration can be found [here](https://library.abb.com/d/9AKK108467A9325). Here we will refer it as **Edge Editor Doc** and the content corresponds to revision G of English version.

### Topology and Services

Field data is read by the Building Edge, mapped to assets, and sent to the ABB Ability Cloud. From there, the assets are retrieved via the Eliona app and automatically created.

![Building edge schema example](https://raw.githubusercontent.com/eliona-smart-building-assistant/open-bos-app/refs/heads/develop/user_guide/building_edge_schema.png "Building Edge schema example")

### Initial Installation
The following configurations need to be done during the initial installation:

#### Local Configuration
Once the Building Edge is powered, it can be accessed via a browser through LAN Port 1 or 2.

Complete the remaining steps according to the **Edge Editor Doc** (Pages 23-32).

#### Cloud Configuration
For the Edge to connect to the cloud, it must be within a network with internet access. Adjust the IP accordingly (see **Edge Editor Doc**, Page 43).

To access the ABB Cloud, a personalized login must be created. After that, follow the steps in **Edge Editor Doc** (Pages 33-40).

#### General Setup
Next, verify and, if necessary, adjust or complete the general settings according to **Edge Editor Doc**, Pages 42-81.

### Integrating Field Data
To retrieve data points from a bus system, a new network must be created under "Field Network" for the corresponding protocol.

The **Ability Cloud** is accessible at: [https://buildings.ability.abb](https://buildings.ability.abb)

Complete the steps as described in **Edge Editor Doc** (Pages 112-123).

### Adding Devices and Data Points
To add devices and data points, the **CSV import** can be used. It is recommended to first export an (empty) CSV file and fill it with data.

### Allowing access via API
To allow Eliona to receive updates from datapoints, every datapoint has to be configured to allow API subscriptions:

1. In Library/Asset Template

2. Select your Asset

3. Click the Edit pen

4. Go to API subscription

5. Enable the one you want.

6. Don't forget to click on Save (Top/Right in the bar Next to API Subscription)

![Allowing API subscription](https://raw.githubusercontent.com/eliona-smart-building-assistant/open-bos-app/refs/heads/develop/user_guide/api_subscription.png "Allowing API subscription")

### Registering the app in OpenBOS

To connect Eliona to your OpenBOS edge, the edge has to be configured to allow API connections. Contact ABB support for more information. You will need to provide them with Eliona's public API URL, which is `https://{your-eliona-domain.io}/apps-public/open-bos`.

Eliona needs the gateway ID, client ID and client secret for authentication.

## Installation

Install the OpenBOS app via the Eliona App Store.

## Configuration

The OpenBOS app requires configuration through Eliona’s settings interface. Below are the steps needed to configure the app.

### Configure the OpenBOS app

Configurations can be created in Eliona under `Settings > Apps > OpenBOS` which opens the app's [Generic Frontend](https://doc.eliona.io/collection/v/eliona-english/manuals/settings/apps). Here you can use the `\configs` with the POST method. Each configuration requires the following data:

| Attribute         | Description                                               |
|-------------------|-----------------------------------------------------------|
| `gwid`            | The ID of the gateway device used in the API requests. |
| `clientID`        | The client ID used for OAuth 2.0 authentication.|
| `clientSecret`    | The client secret used for OAuth 2.0 authentication. |
| `appPublicAPIURL` | URL of this app's public API. Inferred automatically from request. Example: "https://{your-eliona-instance.io}/apps-public/open-bos". |
| `enable`          | Flag to enable or disable fetching from this API. Default: `true`.|
| `refreshInterval` | Interval in hours for resubscribing to OpenBOS. Default: `24`. |
| `requestTimeout`  | API query timeout in seconds. Default: `120`.|
| `active`          | Set to `true` by the app when running and to `false` when app is stopped. Read-only. |
| `projectIDs`      | List of Eliona project IDs for data collection. For each project ID, all smart devices are automatically created as assets in Eliona, with mappings stored in the OpenBOS app. Example: `["42", "99"]`. |

Example full configuration JSON:

```json
{
  "gwid": "1234acbd-3faa-ab32-ab32-21c3876ba",
  "clientID": "4321dcba-3faa-ab32-ab32-21c3876ba",
  "clientSecret": "your-client-secret",
  "appPublicAPIURL": "https://{your-eliona-instance.io}/apps-public/open-bos",
  "enable": true,
  "refreshInterval": 24,
  "requestTimeout": 120,
  "projectIDs": [
    "42",
    "99"
  ]
}

```

Some fields have defaults, so the minimal configuration JSON can be simplified:

```json
{
  "gwid": "1234acbd-3faa-ab32-ab32-21c3876ba",
  "clientID": "4321dcba-3faa-ab32-ab32-21c3876ba",
  "clientSecret": "your-client-secret",
  "projectIDs": [
    "42"
  ]
}

```

## Continuous Asset Creation

Once configured, the app starts Continuous Asset Creation (CAC). Discovered resources are automatically created as assets in Eliona, and user who configured the app is notified via Eliona’s notification system.

### Structuring assets

After creation of assets, the assets are organized in Eliona the same way they are organized in OpenBOS. This structure is kept synchronized, and any movements in Eliona (e.g. renaming asset or moving a room to a different building) will be overwritten with next ontology update.

Only exception are the root assets. Root assets are created in the top-level directory, and correspond to the root of OpenBOS places structure. Typically, there is one Site in this role. Another root asset is the "OpenBOS unassigned" asset, containing unassigned assets and datapoints.

You can rename or move the root assets wherever you want, and the whole ontology will be moved with it.

It is not possible to change GAIs in Eliona for any of assets.

### Asset filtering

In case it's not desired to import all assets from OpenBOS to Eliona, it's possible to write an asset filter that would include only matching assets. This app is able to filter the assets by: id, name and templateID (for both assets and spaces). See [Asset Filter documentation](https://doc.eliona.io/collection/dokumentation/einstellungen/apps/asset-filter) for instructions on writing asset filters.

### Asset types

Asset types are automatically created and synchronized from OpenBOS asset templates. 

| Eliona             | OpenBOS  |
|--------------------|----------|
| Asset type         | Asset template  |
| Attribute - Input  | Datapoint with direction "Feedback"  |
| Attribute - Output | Datapoint with direction "Command" or "CommandAndFeedback" |
| Attribute - Info   | Property  |
| Limits             | Min/Max  |
| Unit               | Unit  |
| Value mapping      | Enums  |

Complex data types from OpenBOS are split into separate attributes in Eliona.

### Orphan datapoints

In case an asset is deleted from OpenBOS and there is still an alarm linked to that datapoint, OpenBOS leaves that datapoint in the ontology. Eliona respects that behaviour, and assigns those datapoints to an "OpenBOS unassigned" asset.

## Alarms

Alarms triggered in OpenBOS are synchronized to Eliona. These are created in Eliona as alarm rules of type "External", and are managed by updates received from OpenBOS -> if an alarm is triggered in OpenBOS, it will be triggered in Eliona as well. Similarly if the alarm is gone.

> *External alarms* are alarms not managed by Eliona. Alarm rules are shown so users can tag and categorize them, but Eliona is not responsible for checking whether the alarm should be raised or not. An external system (in this case OpenBOS) is responsible for triggering the alarms.

If the alarm needs to be acknowledged, users can acknowledge it in Eliona, and this acknowledgement will get synchronized to OpenBOS.

## App status monitoring

Along with asset creation, an asset called "OpenBOS app" is also created. It's purpose is to inform users of the app status -- It signalizes whether the app is running (Asset status -> Active/Inactive) and it's status - the Status attribute. If the app status is not "OK", it signifies that the app might not be functioning properly. If the error state persists, let us know by submitting a bug report.
