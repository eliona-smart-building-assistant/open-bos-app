# OpenBOS User Guide

### Introduction

> The OpenBOS app provides integration and synchronization between Eliona and ABB Open Building Operating System (OpenBOS).

## Overview

This guide provides instructions on configuring, installing, and using the OpenBOS app to manage resources and synchronize data between Eliona and OpenBOS edge.

## Installation

Install the OpenBOS app via the Eliona App Store.

## Configuration

The OpenBOS app requires configuration through Eliona’s settings interface. Below are the steps needed to configure the app.

### Registering the app in OpenBOS

To connect Eliona to your OpenBOS edge, the edge has to be configured to allow API connections. Contact ABB support for more information. You will need to provide them with Eliona's public API URL, which is `https://{your-eliona-domain.io}/apps-public/open-bos`.

Eliona needs the gateway ID, client ID and client secret for authentication.

### Configure the OpenBOS app

Configurations can be created in Eliona under `Settings > Apps > OpenBOS` which opens the app's [Generic Frontend](https://doc.eliona.io/collection/v/eliona-english/manuals/settings/apps). Here you can use the appropriate endpoint with the POST method. Each configuration requires the following data:

| Attribute         | Description                                               |
|-------------------|-----------------------------------------------------------|
| `gwid`            | The ID of the gateway device used in the API requests. |
| `clientID`        | The client ID used for OAuth 2.0 authentication.|
| `clientSecret`    | The client secret used for OAuth 2.0 authentication. |
| `appPublicAPIURL` | URL of this app's public API. Inferred automatically from request. Example: "https://{your-eliona-instance.io}/apps-public/open-bos". |
| `enable`          | Flag to enable or disable fetching from this API. Default: `true`.|
| `refreshInterval` | Interval in seconds for collecting data from API. Default: `60`. |
| `requestTimeout`  | API query timeout in seconds. Default: `120`.|
| `active`          | Set to `true` by the app when running and to `false` when app is stopped. Read-only. |
| `projectIDs`      | List of Eliona project IDs for data collection. For each project ID, all smart devices are automatically created as assets in Eliona, with mappings stored in the KentixONE app. Example: `["42", "99"]`. |

Example configuration JSON:

```json
{
  "gwid": "1234acbd-3faa-ab32-ab32-21c3876ba",
  "clientID": "4321dcba-3faa-ab32-ab32-21c3876ba",
  "clientSecret": "your-client-secret",
  "appPublicAPIURL": "https://{your-eliona-instance.io}/apps-public/open-bos",
  "enable": true,
  "refreshInterval": 60,
  "requestTimeout": 120,
  "projectIDs": [
    "42",
    "99"
  ]
}

```


## Continuous Asset Creation

Once configured, the app starts Continuous Asset Creation (CAC). Discovered resources are automatically created as assets in Eliona, and user who configured the app is notified via Eliona’s notification system.

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

## Alarms

Alarms triggered in OpenBOS are synchronized to Eliona. These are created in Eliona as alarm rules of type "External", and are managed by updates received from OpenBOS -> if an alarm is triggered in OpenBOS, it will be triggered in Eliona as well. Similarly if the alarm is gone.

If the alarm needs to be acknowledged, users can acknowledge it in Eliona, and this acknowledgement will get synchronized to OpenBOS.

### <mark>TODO: Dashboard templates</mark>

The app offers a predefined dashboard that clearly displays the most important information. YOu can create such a dashboard under `Dashboards > Copy Dashboard > From App > OpenBOS`.

### <mark>TODO: Other features</mark>
