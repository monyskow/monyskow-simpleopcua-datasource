# Installation

How to install the Simple OPC-UA datasource plugin in Grafana.

## Requirements

- Grafana 10.4.0 or later
- Access to an OPC-UA server

## From Grafana Plugin Catalog

1. In Grafana, go to **Configuration > Plugins**
2. Search for "Simple OPC-UA"
3. Click **Install**

## Manual Installation

1. Download the latest release from [GitHub Releases](https://github.com/monyskow/monyskow-simpleopcua-datasource/releases)
2. Extract to your Grafana plugins directory (usually `/var/lib/grafana/plugins`)
3. Restart Grafana

## Allow Unsigned Plugin (Development / Local Builds)

Official releases are signed. If you are running a local build or an unreleased version, Grafana must be configured to allow unsigned plugins:

```ini
GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=monyskow-simpleopcua-datasource
```

Set this as an environment variable in your Grafana configuration or `docker-compose` file. The dev `docker-compose.yaml` in this repo already includes it.
