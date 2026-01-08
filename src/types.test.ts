import {
  DEFAULT_QUERY,
  DEFAULT_OPTIONS,
  OpcuaDataSourceOptions,
  OpcuaQuery,
  OpcuaNodeQuery,
  OpcuaBrowseNode,
} from './types';

describe('types', () => {
  describe('DEFAULT_QUERY', () => {
    it('should have empty nodes array', () => {
      expect(DEFAULT_QUERY.nodes).toEqual([]);
    });
  });

  describe('DEFAULT_OPTIONS', () => {
    it('should have correct default values', () => {
      expect(DEFAULT_OPTIONS.endpoint).toBe('');
      expect(DEFAULT_OPTIONS.securityPolicy).toBe('None');
      expect(DEFAULT_OPTIONS.securityMode).toBe('None');
      expect(DEFAULT_OPTIONS.authMethod).toBe('anonymous');
      expect(DEFAULT_OPTIONS.timeout).toBe(10);
    });
  });

  describe('OpcuaDataSourceOptions', () => {
    it('should accept valid options', () => {
      const options: OpcuaDataSourceOptions = {
        endpoint: 'opc.tcp://localhost:4840',
        securityPolicy: 'Basic256Sha256',
        securityMode: 'SignAndEncrypt',
        authMethod: 'userpass',
        timeout: 30,
      };

      expect(options.endpoint).toBe('opc.tcp://localhost:4840');
      expect(options.securityPolicy).toBe('Basic256Sha256');
      expect(options.securityMode).toBe('SignAndEncrypt');
      expect(options.authMethod).toBe('userpass');
      expect(options.timeout).toBe(30);
    });
  });

  describe('OpcuaQuery', () => {
    it('should accept valid query', () => {
      const query: OpcuaQuery = {
        refId: 'A',
        nodes: [
          {
            nodeId: 'ns=2;s=Temperature',
            displayName: 'Temperature',
            alias: 'Temp',
          },
        ],
      };

      expect(query.refId).toBe('A');
      expect(query.nodes.length).toBe(1);
      expect(query.nodes[0].nodeId).toBe('ns=2;s=Temperature');
    });
  });

  describe('OpcuaNodeQuery', () => {
    it('should require nodeId and displayName', () => {
      const node: OpcuaNodeQuery = {
        nodeId: 'ns=2;s=Test',
        displayName: 'Test Node',
      };

      expect(node.nodeId).toBe('ns=2;s=Test');
      expect(node.displayName).toBe('Test Node');
      expect(node.alias).toBeUndefined();
    });

    it('should allow optional alias', () => {
      const node: OpcuaNodeQuery = {
        nodeId: 'ns=2;s=Test',
        displayName: 'Test Node',
        alias: 'TestAlias',
      };

      expect(node.alias).toBe('TestAlias');
    });
  });

  describe('OpcuaBrowseNode', () => {
    it('should have all required properties', () => {
      const browseNode: OpcuaBrowseNode = {
        nodeId: 'ns=2;s=MyNode',
        displayName: 'My Node',
        browseName: 'MyNode',
        nodeClass: 'Variable',
        hasChildren: false,
      };

      expect(browseNode.nodeId).toBe('ns=2;s=MyNode');
      expect(browseNode.displayName).toBe('My Node');
      expect(browseNode.browseName).toBe('MyNode');
      expect(browseNode.nodeClass).toBe('Variable');
      expect(browseNode.hasChildren).toBe(false);
    });

    it('should identify objects with hasChildren true', () => {
      const objectNode: OpcuaBrowseNode = {
        nodeId: 'ns=2;s=MyFolder',
        displayName: 'My Folder',
        browseName: 'MyFolder',
        nodeClass: 'Object',
        hasChildren: true,
      };

      expect(objectNode.nodeClass).toBe('Object');
      expect(objectNode.hasChildren).toBe(true);
    });
  });
});
