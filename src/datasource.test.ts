import { DataSourceInstanceSettings, CoreApp } from '@grafana/data';
import { OpcuaDataSource } from './datasource';
import { OpcuaDataSourceOptions, OpcuaQuery } from './types';

// Mock @grafana/runtime
jest.mock('@grafana/runtime', () => {
  // Get actual DataSourceWithBackend from @grafana/runtime
  const actual = jest.requireActual('@grafana/runtime');
  return {
    ...actual,
    getTemplateSrv: () => ({
      replace: (str: string) => str.replace(/\$\{(\w+)\}/g, '$1_value'),
    }),
  };
});

describe('OpcuaDataSource', () => {
  const instanceSettings: DataSourceInstanceSettings<OpcuaDataSourceOptions> = {
    id: 1,
    uid: 'test-uid',
    type: 'monyskow-simpleopcua-datasource',
    name: 'Test OPC-UA',
    meta: {} as any,
    readOnly: false,
    jsonData: {
      endpoint: 'opc.tcp://localhost:4840',
      securityPolicy: 'None',
      securityMode: 'None',
      authMethod: 'anonymous',
      timeout: 10,
    },
    access: 'proxy',
  };

  let datasource: OpcuaDataSource;

  beforeEach(() => {
    datasource = new OpcuaDataSource(instanceSettings);
  });

  describe('getDefaultQuery', () => {
    it('should return default query with empty nodes array', () => {
      const result = datasource.getDefaultQuery(CoreApp.PanelEditor);

      expect(result).toEqual({
        nodes: [],
      });
    });
  });

  describe('filterQuery', () => {
    it('should return false for empty nodes array', () => {
      const query: OpcuaQuery = {
        refId: 'A',
        nodes: [],
      };

      expect(datasource.filterQuery(query)).toBe(false);
    });

    it('should return true for query with nodes', () => {
      const query: OpcuaQuery = {
        refId: 'A',
        nodes: [
          {
            nodeId: 'ns=2;s=Temperature',
            displayName: 'Temperature',
          },
        ],
      };

      expect(datasource.filterQuery(query)).toBe(true);
    });

    it('should return false for undefined nodes', () => {
      const query = {
        refId: 'A',
      } as OpcuaQuery;

      expect(datasource.filterQuery(query)).toBe(false);
    });
  });

  describe('applyTemplateVariables', () => {
    it('should replace template variables in node IDs', () => {
      const query: OpcuaQuery = {
        refId: 'A',
        nodes: [
          {
            nodeId: 'ns=2;s=${machine}/Temperature',
            displayName: 'Temperature',
            alias: '${machine}_temp',
          },
        ],
      };

      const result = datasource.applyTemplateVariables(query, {});

      // Mock replaces ${var} with var_value
      expect(result.nodes).toBeDefined();
      expect(result.nodes.length).toBe(1);
      expect(result.nodes[0].nodeId).toBe('ns=2;s=machine_value/Temperature');
      expect(result.nodes[0].alias).toBe('machine_value_temp');
    });

    it('should handle empty nodes array', () => {
      const query: OpcuaQuery = {
        refId: 'A',
        nodes: [],
      };

      const result = datasource.applyTemplateVariables(query, {});

      expect(result.nodes).toEqual([]);
    });

    it('should handle undefined alias', () => {
      const query: OpcuaQuery = {
        refId: 'A',
        nodes: [
          {
            nodeId: 'ns=2;s=Test',
            displayName: 'Test',
          },
        ],
      };

      const result = datasource.applyTemplateVariables(query, {});

      expect(result.nodes[0].alias).toBeUndefined();
    });
  });
});
