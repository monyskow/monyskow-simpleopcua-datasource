import React from 'react';
import { render, RenderOptions } from '@testing-library/react';

/**
 * Custom render function that wraps components with necessary providers
 */
const customRender = (ui: React.ReactElement, options?: Omit<RenderOptions, 'wrapper'>) => render(ui, { ...options });

// Re-export everything from testing-library
export * from '@testing-library/react';

// Override render method
export { customRender as render };

/**
 * Mock data source for testing
 */
export const mockDatasource = {
  id: 1,
  uid: 'test-datasource',
  type: 'monyskow-simpleopcua-datasource',
  name: 'Test OPC-UA',
  meta: {
    id: 'monyskow-simpleopcua-datasource',
    name: 'Simple OPC-UA',
    type: 'datasource',
    info: {
      author: { name: 'monyskow' },
      description: 'OPC-UA data source',
      logos: { small: '', large: '' },
      links: [],
      screenshots: [],
      version: '1.0.0',
      updated: '',
    },
    module: '',
    baseUrl: '',
  },
  jsonData: {
    endpoint: 'opc.tcp://localhost:4840',
    securityPolicy: 'None' as const,
    securityMode: 'None' as const,
    authMethod: 'anonymous' as const,
    timeout: 10,
  },
};

/**
 * Mock query for testing
 */
export const mockQuery = {
  refId: 'A',
  nodes: [
    {
      nodeId: 'ns=2;s=Temperature',
      displayName: 'Temperature',
      alias: 'Temp',
    },
  ],
};

/**
 * Mock browse node for testing
 */
export const mockBrowseNode = {
  nodeId: 'ns=2;s=TestNode',
  displayName: 'Test Node',
  browseName: 'TestNode',
  nodeClass: 'Variable',
  hasChildren: false,
};
