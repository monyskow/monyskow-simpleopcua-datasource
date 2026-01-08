import { DataSourcePlugin } from '@grafana/data';
import { OpcuaDataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor/ConfigEditor';
import { QueryEditor } from './components/QueryEditor/QueryEditor';
import { OpcuaDataSourceOptions, OpcuaQuery } from './types';

/**
 * Plugin registration
 *
 * Registers the OPC-UA data source with Grafana, connecting:
 * - The DataSource class for data fetching
 * - The ConfigEditor for data source configuration UI
 * - The QueryEditor for panel query building UI
 */
export const plugin = new DataSourcePlugin<
  OpcuaDataSource,
  OpcuaQuery,
  OpcuaDataSourceOptions
>(OpcuaDataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
