import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

/**
 * Authentication methods supported by the OPC-UA data source
 */
export type AuthMethod = 'anonymous' | 'userpass' | 'certificate';

/**
 * OPC-UA Security Policies
 */
export type SecurityPolicy =
  | 'None'
  | 'Basic128Rsa15'
  | 'Basic256'
  | 'Basic256Sha256'
  | 'Aes128_Sha256_RsaOaep'
  | 'Aes256_Sha256_RsaPss';

/**
 * OPC-UA Security Modes
 */
export type SecurityMode = 'None' | 'Sign' | 'SignAndEncrypt';

/**
 * Data source configuration stored in jsonData (not encrypted)
 */
export interface OpcuaDataSourceOptions extends DataSourceJsonData {
  /** OPC-UA server endpoint (e.g., opc.tcp://localhost:4840) */
  endpoint: string;
  /** Security policy for encryption */
  securityPolicy: SecurityPolicy;
  /** Security mode (None, Sign, or SignAndEncrypt) */
  securityMode: SecurityMode;
  /** Authentication method */
  authMethod: AuthMethod;
  /** Connection timeout in seconds */
  timeout: number;
}

/**
 * Secure configuration stored in secureJsonData (encrypted at rest)
 */
export interface OpcuaSecureJsonData {
  /** Username for username/password authentication */
  username?: string;
  /** Password for username/password authentication */
  password?: string;
  /** PEM-encoded certificate for certificate authentication */
  certificate?: string;
  /** PEM-encoded private key for certificate authentication */
  privateKey?: string;
  /** Auto-generated client certificate for secure connections (PEM) */
  clientCert?: string;
  /** Auto-generated client private key for secure connections (PEM) */
  clientKey?: string;
}

/**
 * Query model for the OPC-UA data source
 */
export interface OpcuaQuery extends DataQuery {
  /** List of OPC-UA nodes to read */
  nodes: OpcuaNodeQuery[];
}

/**
 * Individual node query configuration
 */
export interface OpcuaNodeQuery {
  /** OPC-UA Node ID (e.g., "ns=2;s=MyVariable") */
  nodeId: string;
  /** Human-readable display name */
  displayName: string;
  /** Optional alias for the field name in the response */
  alias?: string;
}

/**
 * Node information returned from browsing
 */
export interface OpcuaBrowseNode {
  /** OPC-UA Node ID */
  nodeId: string;
  /** Display name of the node */
  displayName: string;
  /** Browse name */
  browseName: string;
  /** Node class (Object, Variable, Method, etc.) */
  nodeClass: string;
  /** Whether this node has children that can be browsed */
  hasChildren: boolean;
}

/**
 * Default query configuration
 */
export const DEFAULT_QUERY: Partial<OpcuaQuery> = {
  nodes: [],
};

/**
 * Default data source options
 */
export const DEFAULT_OPTIONS: Partial<OpcuaDataSourceOptions> = {
  endpoint: '',
  securityPolicy: 'None',
  securityMode: 'None',
  authMethod: 'anonymous',
  timeout: 10,
};
