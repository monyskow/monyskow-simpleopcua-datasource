import {
  DataSourceInstanceSettings,
  CoreApp,
  ScopedVars,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import {
  OpcuaDataSourceOptions,
  OpcuaQuery,
  OpcuaBrowseNode,
  DEFAULT_QUERY,
} from './types';

/**
 * OPC-UA Data Source
 *
 * Connects to OPC-UA servers to read industrial data.
 * Uses a Go backend for OPC-UA protocol handling.
 */
export class OpcuaDataSource extends DataSourceWithBackend<
  OpcuaQuery,
  OpcuaDataSourceOptions
> {
  constructor(
    instanceSettings: DataSourceInstanceSettings<OpcuaDataSourceOptions>
  ) {
    super(instanceSettings);
  }

  /**
   * Apply template variables to query before sending to backend
   */
  applyTemplateVariables(
    query: OpcuaQuery,
    scopedVars: ScopedVars
  ): OpcuaQuery {
    const templateSrv = getTemplateSrv();

    return {
      ...query,
      nodes: (query.nodes || []).map((node) => ({
        ...node,
        nodeId: templateSrv.replace(node.nodeId, scopedVars),
        alias: node.alias
          ? templateSrv.replace(node.alias, scopedVars)
          : undefined,
      })),
    };
  }

  /**
   * Returns the default query for new panels
   */
  getDefaultQuery(_app: CoreApp): Partial<OpcuaQuery> {
    return DEFAULT_QUERY;
  }

  /**
   * Filter queries - exclude empty queries from execution
   */
  filterQuery(query: OpcuaQuery): boolean {
    return Boolean(query.nodes && query.nodes.length > 0);
  }

  /**
   * Browse OPC-UA nodes via the backend resource API
   *
   * @param nodeId - Parent node ID to browse (default: Objects folder)
   * @returns Promise resolving to array of child nodes
   */
  async browseNodes(nodeId?: string): Promise<OpcuaBrowseNode[]> {
    const params = nodeId ? `?nodeId=${encodeURIComponent(nodeId)}` : '';
    return this.getResource<OpcuaBrowseNode[]>(`browse${params}`);
  }

  /**
   * Get available endpoints from the OPC-UA server
   *
   * @returns Promise resolving to array of endpoint descriptions
   */
  async getEndpoints(): Promise<EndpointDescription[]> {
    return this.getResource<EndpointDescription[]>('endpoints');
  }
}

/**
 * OPC-UA Endpoint Description (simplified)
 */
export interface EndpointDescription {
  endpointUrl: string;
  securityMode: string;
  securityPolicy: string;
  userIdentityTokens: string[];
}
