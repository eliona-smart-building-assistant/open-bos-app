# OpenBOS User Guide

### Introduction

> The OpenBOS app provides integration and synchronization between Eliona and ABB OpenBOS.

## Overview

This guide provides instructions on configuring, installing, and using the OpenBOS app to manage resources and synchronize data between Eliona and OpenBOS services.

## Installation

Install the OpenBOS app via the Eliona App Store.

## Configuration

The OpenBOS app requires configuration through Eliona’s settings interface. Below are the general steps and details needed to configure the app effectively.

### Registering the app in OpenBOS Service

Create credentials in OpenBOS Service to connect the OpenBOS services from Eliona. All required credentials are listed below in the [configuration section](#configure-the-open-bos-app).  

<mark>TODO: Describe the steps where you can get or create the necessary credentials.</mark> 

### Configure the OpenBOS app 

Configurations can be created in Eliona under `Apps > OpenBOS > Settings` which opens the app's [Generic Frontend](https://doc.eliona.io/collection/v/eliona-english/manuals/settings/apps). Here you can use the appropriate endpoint with the POST method. Each configuration requires the following data:

| Attribute         | Description                                                                     |
|-------------------|---------------------------------------------------------------------------------|
| `baseURL`         | URL of the OpenBOS services.                                                   |
| `clientSecrets`   | Client secrets obtained from the OpenBOS service.                              |
| `assetFilter`     | Filtering asset during [Continuous Asset Creation](#continuous-asset-creation). |
| `enable`          | Flag to enable or disable this configuration.                                   |
| `refreshInterval` | Interval in seconds for data synchronization.                                   |
| `requestTimeout`  | API query timeout in seconds.                                                   |
| `projectIDs`      | List of Eliona project IDs for data collection.                                 |

Example configuration JSON:

```json
{
  "baseURL": "http://service/v1",
  "clientSecrets": "random-cl13nt-s3cr3t",
  "filter": "",
  "enable": true,
  "refreshInterval": 60,
  "requestTimeout": 120,
  "projectIDs": [
    "10"
  ]
}
```

## Continuous Asset Creation

Once configured, the app starts Continuous Asset Creation (CAC). Discovered resources are automatically created as assets in Eliona, and users are notified via Eliona’s notification system.

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

## Additional Features

<mark>TODO: Describe all other features of the app.</mark>

### Dashboard templates

The app offers a predefined dashboard that clearly displays the most important information. YOu can create such a dashboard under `Dashboards > Copy Dashboard > From App > OpenBOS`.

### <mark>TODO: Other features</mark>
