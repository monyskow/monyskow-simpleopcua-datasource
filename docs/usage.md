# Usage

How to build queries and use template variables with the Simple OPC-UA datasource.

## Building Queries

1. Add a new panel and select the OPC-UA data source
2. Click "Browse Nodes" to explore the OPC-UA address space
3. Click the + icon next to any Variable node to add it to your query
4. Optionally, set an alias for each node

## Using Template Variables

You can use Grafana template variables in Node IDs:

```
ns=2;s=${machine}/Temperature
```
