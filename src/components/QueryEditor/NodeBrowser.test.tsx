import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { NodeBrowser } from './NodeBrowser';
import { OpcuaDataSource } from '../../datasource';
import { OpcuaBrowseNode } from '../../types';

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  getTemplateSrv: () => ({ replace: (s: string) => s }),
}));

// Override the scaffolded react-inlinesvg mock so that onClick and other
// event props forwarded by @grafana/ui's Icon component are actually attached
// to the rendered SVG element (the scaffolded mock drops all props except src).
jest.mock('react-inlinesvg', () => {
  const SVG_FILE_NAME_REGEX = /(.+)\/(.+)\.svg$/;
  // Forward event props (onClick etc.) but drop non-DOM props like innerRef/loader
  const MockSVG = ({
    src,
    innerRef: _innerRef,
    loader: _loader,
    ...rest
  }: {
    src: string;
    innerRef?: unknown;
    loader?: unknown;
    [key: string]: unknown;
  }) => {
    const testId = src.replace(SVG_FILE_NAME_REGEX, '$2');
    return (
      <svg
        xmlns="http://www.w3.org/2000/svg"
        data-testid={testId}
        viewBox="0 0 24 24"
        {...(rest as React.SVGProps<SVGSVGElement>)}
      />
    );
  };
  return { __esModule: true, default: MockSVG };
});

function makeDatasource(browseNodes: jest.Mock) {
  return { browseNodes } as unknown as OpcuaDataSource;
}

const variableNode: OpcuaBrowseNode = {
  nodeId: 'ns=2;s=Temperature',
  displayName: 'Temperature',
  browseName: 'Temperature',
  nodeClass: 'Variable',
  hasChildren: false,
};

const folderNode: OpcuaBrowseNode = {
  nodeId: 'ns=2;s=Folder',
  displayName: 'MyFolder',
  browseName: 'MyFolder',
  nodeClass: 'Object',
  hasChildren: true,
};

describe('NodeBrowser', () => {
  describe('loading state', () => {
    it('shows loading spinner while fetching root nodes', () => {
      // browseNodes never resolves during this test
      const browseNodes = jest.fn(() => new Promise<OpcuaBrowseNode[]>(() => {}));
      render(<NodeBrowser datasource={makeDatasource(browseNodes)} onSelect={jest.fn()} />);
      expect(screen.getByText(/Loading nodes/)).toBeInTheDocument();
    });
  });

  describe('populated tree', () => {
    it('renders root nodes after loading', async () => {
      const browseNodes = jest.fn().mockResolvedValue([variableNode]);
      render(<NodeBrowser datasource={makeDatasource(browseNodes)} onSelect={jest.fn()} />);

      await waitFor(() => {
        expect(screen.getByText('Temperature')).toBeInTheDocument();
      });
    });

    it('calls datasource.browseNodes with no argument for root', async () => {
      const browseNodes = jest.fn().mockResolvedValue([variableNode]);
      render(<NodeBrowser datasource={makeDatasource(browseNodes)} onSelect={jest.fn()} />);

      await waitFor(() => expect(browseNodes).toHaveBeenCalledTimes(1));
      expect(browseNodes).toHaveBeenCalledWith(/* no args */);
    });

    it('shows plus-circle add icon on variable nodes', async () => {
      const browseNodes = jest.fn().mockResolvedValue([variableNode]);
      render(<NodeBrowser datasource={makeDatasource(browseNodes)} onSelect={jest.fn()} />);

      await waitFor(() => expect(screen.getByText('Temperature')).toBeInTheDocument());
      // Icon renders an SVG with data-testid matching the icon name
      expect(screen.getByTestId('plus-circle')).toBeInTheDocument();
    });

    it('calls onSelect with nodeId and displayName when add icon is clicked', async () => {
      const browseNodes = jest.fn().mockResolvedValue([variableNode]);
      const onSelect = jest.fn();
      render(<NodeBrowser datasource={makeDatasource(browseNodes)} onSelect={onSelect} />);

      await waitFor(() => expect(screen.getByTestId('plus-circle')).toBeInTheDocument());
      fireEvent.click(screen.getByTestId('plus-circle'));

      expect(onSelect).toHaveBeenCalledWith('ns=2;s=Temperature', 'Temperature');
    });
  });

  describe('folder expansion', () => {
    it('expands a folder node and loads children on click', async () => {
      const childNode: OpcuaBrowseNode = {
        nodeId: 'ns=2;s=Child',
        displayName: 'ChildNode',
        browseName: 'ChildNode',
        nodeClass: 'Variable',
        hasChildren: false,
      };
      const browseNodes = jest.fn().mockResolvedValueOnce([folderNode]).mockResolvedValueOnce([childNode]);

      render(<NodeBrowser datasource={makeDatasource(browseNodes)} onSelect={jest.fn()} />);

      await waitFor(() => expect(screen.getByText('MyFolder')).toBeInTheDocument());

      fireEvent.click(screen.getByRole('treeitem', { name: 'MyFolder' }));

      await waitFor(() => {
        expect(screen.getByText('ChildNode')).toBeInTheDocument();
      });
      expect(browseNodes).toHaveBeenCalledWith('ns=2;s=Folder');
    });

    it('collapses an expanded folder on second click', async () => {
      const childNode: OpcuaBrowseNode = {
        nodeId: 'ns=2;s=Child',
        displayName: 'ChildNode',
        browseName: 'ChildNode',
        nodeClass: 'Variable',
        hasChildren: false,
      };
      const browseNodes = jest.fn().mockResolvedValueOnce([folderNode]).mockResolvedValueOnce([childNode]);

      render(<NodeBrowser datasource={makeDatasource(browseNodes)} onSelect={jest.fn()} />);

      await waitFor(() => expect(screen.getByText('MyFolder')).toBeInTheDocument());

      // Expand
      fireEvent.click(screen.getByRole('treeitem', { name: 'MyFolder' }));
      await waitFor(() => expect(screen.getByText('ChildNode')).toBeInTheDocument());

      // Collapse
      fireEvent.click(screen.getByRole('treeitem', { name: 'MyFolder' }));
      expect(screen.queryByText('ChildNode')).not.toBeInTheDocument();
    });
  });

  describe('error state', () => {
    it('shows error alert when browseNodes rejects', async () => {
      const browseNodes = jest.fn().mockRejectedValue(new Error('Connection refused'));
      render(<NodeBrowser datasource={makeDatasource(browseNodes)} onSelect={jest.fn()} />);

      await waitFor(() => {
        expect(screen.getByText('Connection refused')).toBeInTheDocument();
      });
    });
  });

  describe('empty state', () => {
    it('shows empty-state alert when server returns no nodes', async () => {
      const browseNodes = jest.fn().mockResolvedValue([]);
      render(<NodeBrowser datasource={makeDatasource(browseNodes)} onSelect={jest.fn()} />);

      await waitFor(() => {
        expect(screen.getByText(/No nodes found/)).toBeInTheDocument();
      });
    });
  });
});
