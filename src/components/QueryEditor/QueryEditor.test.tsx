import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryEditor } from './QueryEditor';
import { OpcuaDataSource } from '../../datasource';
import { OpcuaQuery } from '../../types';

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  getTemplateSrv: () => ({ replace: (s: string) => s }),
}));

// Modal from @grafana/ui uses portals that don't render well in jsdom; replace with
// a simple div so we can test its children without portal issues.
jest.mock('@grafana/ui', () => {
  const actual = jest.requireActual('@grafana/ui');
  const MockModal = ({
    isOpen,
    children,
    title,
  }: {
    isOpen: boolean;
    children: React.ReactNode;
    title: string;
    onDismiss: () => void;
  }) =>
    isOpen ? (
      <div role="dialog" aria-label={title}>
        {children}
      </div>
    ) : null;
  return { ...actual, Modal: MockModal };
});

// NodeBrowser makes async calls to datasource.browseNodes; stub it out for QueryEditor tests.
jest.mock('./NodeBrowser', () => ({
  NodeBrowser: ({ onSelect }: { onSelect: (nodeId: string, displayName: string) => void }) => (
    <button onClick={() => onSelect('ns=2;s=FromBrowser', 'Browser Node')}>Select From Browser</button>
  ),
}));

function makeQuery(overrides: Partial<OpcuaQuery> = {}): OpcuaQuery {
  return { refId: 'A', nodes: [], ...overrides };
}

function makeDatasource() {
  return { browseNodes: jest.fn().mockResolvedValue([]) } as unknown as OpcuaDataSource;
}

function renderEditor(queryOverrides: Partial<OpcuaQuery> = {}, onChange = jest.fn(), onRunQuery = jest.fn()) {
  const query = makeQuery(queryOverrides);
  const datasource = makeDatasource();
  return {
    onChange,
    onRunQuery,
    ...render(
      <QueryEditor
        query={query}
        onChange={onChange}
        onRunQuery={onRunQuery}
        datasource={datasource}
        data={undefined as any}
        range={undefined as any}
        history={[]}
        queries={[]}
        app={undefined as any}
      />
    ),
  };
}

describe('QueryEditor', () => {
  describe('initial render', () => {
    it('shows Browse Nodes button', () => {
      renderEditor();
      expect(screen.getByRole('button', { name: 'Browse Nodes' })).toBeInTheDocument();
    });

    it('shows Add Manual button', () => {
      renderEditor();
      expect(screen.getByRole('button', { name: 'Add Manual Node' })).toBeInTheDocument();
    });

    it('shows empty state message when no nodes are configured', () => {
      renderEditor();
      expect(screen.getByText(/No nodes configured/)).toBeInTheDocument();
    });

    it('shows configured nodes', () => {
      renderEditor({
        nodes: [{ nodeId: 'ns=2;s=Temperature', displayName: 'Temperature' }],
      });
      expect(screen.getByDisplayValue('ns=2;s=Temperature')).toBeInTheDocument();
    });
  });

  describe('Add Manual button', () => {
    it('calls onChange with a new empty node when clicked', () => {
      const onChange = jest.fn();
      renderEditor({}, onChange);

      fireEvent.click(screen.getByRole('button', { name: 'Add Manual Node' }));

      expect(onChange).toHaveBeenCalledWith(
        expect.objectContaining({
          nodes: [{ nodeId: '', displayName: 'New Node' }],
        })
      );
    });

    it('calls onRunQuery when Add Manual is clicked', () => {
      const onRunQuery = jest.fn();
      renderEditor({}, jest.fn(), onRunQuery);

      fireEvent.click(screen.getByRole('button', { name: 'Add Manual Node' }));

      expect(onRunQuery).toHaveBeenCalled();
    });
  });

  describe('Node ID input interaction', () => {
    it('calls onChange with updated nodeId when input changes', () => {
      const onChange = jest.fn();
      renderEditor({ nodes: [{ nodeId: 'ns=2;s=Old', displayName: 'Old' }] }, onChange);

      const input = screen.getByLabelText('Node ID 1');
      fireEvent.change(input, { target: { value: 'ns=2;s=New' } });

      expect(onChange).toHaveBeenCalledWith(
        expect.objectContaining({
          nodes: [expect.objectContaining({ nodeId: 'ns=2;s=New' })],
        })
      );
    });

    it('calls onRunQuery on blur of node ID input', () => {
      const onRunQuery = jest.fn();
      renderEditor({ nodes: [{ nodeId: 'ns=2;s=Test', displayName: 'Test' }] }, jest.fn(), onRunQuery);

      fireEvent.blur(screen.getByLabelText('Node ID 1'));

      expect(onRunQuery).toHaveBeenCalled();
    });
  });

  describe('Alias input interaction', () => {
    it('calls onChange with updated alias when alias input changes', () => {
      const onChange = jest.fn();
      renderEditor({ nodes: [{ nodeId: 'ns=2;s=Temp', displayName: 'Temperature' }] }, onChange);

      const aliasInput = screen.getByLabelText('Alias 1');
      fireEvent.change(aliasInput, { target: { value: 'TempAlias' } });

      expect(onChange).toHaveBeenCalledWith(
        expect.objectContaining({
          nodes: [expect.objectContaining({ alias: 'TempAlias' })],
        })
      );
    });
  });

  describe('Remove node', () => {
    it('calls onChange with node removed when trash icon clicked', () => {
      const onChange = jest.fn();
      renderEditor({ nodes: [{ nodeId: 'ns=2;s=A', displayName: 'A' }] }, onChange);

      // IconButton renders aria-label="Remove node" (ignores the index suffix we pass in)
      fireEvent.click(screen.getByLabelText('Remove node'));

      expect(onChange).toHaveBeenCalledWith(
        expect.objectContaining({
          nodes: [],
        })
      );
    });

    it('calls onRunQuery after removing a node', () => {
      const onRunQuery = jest.fn();
      renderEditor({ nodes: [{ nodeId: 'ns=2;s=A', displayName: 'A' }] }, jest.fn(), onRunQuery);

      fireEvent.click(screen.getByLabelText('Remove node'));

      expect(onRunQuery).toHaveBeenCalled();
    });
  });

  describe('Node browser integration', () => {
    it('opens node browser modal when Browse Nodes is clicked', () => {
      renderEditor();
      fireEvent.click(screen.getByRole('button', { name: 'Browse Nodes' }));
      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });

    it('adds selected node from browser and closes modal', () => {
      const onChange = jest.fn();
      renderEditor({}, onChange);

      fireEvent.click(screen.getByRole('button', { name: 'Browse Nodes' }));
      fireEvent.click(screen.getByRole('button', { name: 'Select From Browser' }));

      expect(onChange).toHaveBeenCalledWith(
        expect.objectContaining({
          nodes: [{ nodeId: 'ns=2;s=FromBrowser', displayName: 'Browser Node' }],
        })
      );
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });
  });
});
