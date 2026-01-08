import React, { useState } from 'react';
import { QueryEditorProps } from '@grafana/data';
import {
  Button,
  InlineFieldRow,
  InlineField,
  Input,
  IconButton,
  Modal,
} from '@grafana/ui';
import { OpcuaDataSource } from '../../datasource';
import {
  OpcuaDataSourceOptions,
  OpcuaQuery,
  OpcuaNodeQuery,
} from '../../types';
import { NodeBrowser } from './NodeBrowser';

type Props = QueryEditorProps<
  OpcuaDataSource,
  OpcuaQuery,
  OpcuaDataSourceOptions
>;

/**
 * Query editor for the OPC-UA data source
 *
 * Allows users to:
 * - Browse OPC-UA nodes using the NodeBrowser
 * - Manually add nodes by entering Node IDs
 * - Configure aliases for each node
 * - Remove nodes from the query
 */
export const QueryEditor: React.FC<Props> = ({
  query,
  onChange,
  onRunQuery,
  datasource,
}) => {
  const [showBrowser, setShowBrowser] = useState(false);

  const nodes = query.nodes || [];

  const onAddNode = (node: OpcuaNodeQuery) => {
    onChange({
      ...query,
      nodes: [...nodes, node],
    });
    onRunQuery();
  };

  const onRemoveNode = (index: number) => {
    const newNodes = [...nodes];
    newNodes.splice(index, 1);
    onChange({
      ...query,
      nodes: newNodes,
    });
    onRunQuery();
  };

  const onUpdateNode = (
    index: number,
    field: keyof OpcuaNodeQuery,
    value: string
  ) => {
    const newNodes = [...nodes];
    newNodes[index] = { ...newNodes[index], [field]: value };
    onChange({
      ...query,
      nodes: newNodes,
    });
  };

  const onNodeSelect = (nodeId: string, displayName: string) => {
    onAddNode({ nodeId, displayName });
    setShowBrowser(false);
  };

  const handleBlur = () => {
    onRunQuery();
  };

  return (
    <div>
      <InlineFieldRow>
        <Button
          icon="plus"
          variant="secondary"
          size="sm"
          onClick={() => setShowBrowser(true)}
          aria-label="Browse Nodes"
        >
          Browse Nodes
        </Button>
        <Button
          icon="keyboard"
          variant="secondary"
          size="sm"
          onClick={() => onAddNode({ nodeId: '', displayName: 'New Node' })}
          style={{ marginLeft: 8 }}
          aria-label="Add Manual Node"
        >
          Add Manual
        </Button>
      </InlineFieldRow>

      {nodes.length === 0 && (
        <div style={{ marginTop: 16, color: '#8e8e8e' }}>
          No nodes configured. Use &quot;Browse Nodes&quot; to select OPC-UA
          variables or &quot;Add Manual&quot; to enter Node IDs directly.
        </div>
      )}

      {nodes.map((node, index) => (
        <InlineFieldRow key={index} style={{ marginTop: 8 }}>
          <InlineField label="Node ID" labelWidth={10}>
            <Input
              width={40}
              value={node.nodeId}
              onChange={(e) =>
                onUpdateNode(index, 'nodeId', e.currentTarget.value)
              }
              onBlur={handleBlur}
              placeholder="ns=2;s=MyVariable"
              aria-label={`Node ID ${index + 1}`}
            />
          </InlineField>
          <InlineField label="Alias" labelWidth={8}>
            <Input
              width={20}
              value={node.alias || ''}
              onChange={(e) =>
                onUpdateNode(index, 'alias', e.currentTarget.value)
              }
              onBlur={handleBlur}
              placeholder={node.displayName || 'Field name'}
              aria-label={`Alias ${index + 1}`}
            />
          </InlineField>
          <IconButton
            name="trash-alt"
            tooltip="Remove node"
            onClick={() => onRemoveNode(index)}
            aria-label={`Remove node ${index + 1}`}
          />
        </InlineFieldRow>
      ))}

      <Modal
        title="Browse OPC-UA Nodes"
        isOpen={showBrowser}
        onDismiss={() => setShowBrowser(false)}
      >
        <NodeBrowser datasource={datasource} onSelect={onNodeSelect} />
      </Modal>
    </div>
  );
};
