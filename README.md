# Simple OPC-UA Data Source for Grafana

A Grafana data source plugin for connecting to OPC-UA servers. Read industrial data directly in Grafana dashboards.

## What It Does

- Easy configuration with endpoint URL, security settings, and authentication
- Multiple auth methods: Anonymous, Username/Password, and Certificate
- Graphical node browser for exploring the OPC-UA address space
- Full support for Grafana template variables in Node IDs
- Built-in connection testing via Save & Test

## Requirements

- Grafana 10.4.0 or later
- Access to an OPC-UA server

## Documentation

| Doc                                                                                                              | Description                                            |
| ---------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| [Installation](https://github.com/monyskow/monyskow-simpleopcua-datasource/blob/main/docs/installation.md)       | From catalog, manual install, unsigned plugin setup    |
| [Configuration](https://github.com/monyskow/monyskow-simpleopcua-datasource/blob/main/docs/configuration.md)     | Connection settings, security policies, authentication |
| [Usage](https://github.com/monyskow/monyskow-simpleopcua-datasource/blob/main/docs/usage.md)                     | Building queries, template variables                   |
| [Development](https://github.com/monyskow/monyskow-simpleopcua-datasource/blob/main/docs/development.md)         | Build commands, project layout, test servers           |
| [Testing](https://github.com/monyskow/monyskow-simpleopcua-datasource/blob/main/docs/testing.md)                 | Manual testing guide, E2E tests, CI matrix             |
| [Troubleshooting](https://github.com/monyskow/monyskow-simpleopcua-datasource/blob/main/docs/troubleshooting.md) | Known issues, Grafana 13 quirks, EOF errors            |

## License

Apache License 2.0. See [LICENSE](https://github.com/monyskow/monyskow-simpleopcua-datasource/blob/main/LICENSE) for details.

## Support

- [GitHub Issues](https://github.com/monyskow/monyskow-simpleopcua-datasource/issues)
