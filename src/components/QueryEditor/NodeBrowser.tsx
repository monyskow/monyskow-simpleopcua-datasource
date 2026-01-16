import React, { useEffect, useState, useCallback } from 'react';
import { Icon, Spinner, Alert } from '@grafana/ui';
import { css } from '@emotion/css';
import { OpcuaDataSource } from '../../datasource';
import { OpcuaBrowseNode } from '../../types';

interface Props {
  datasource: OpcuaDataSource;
  onSelect: (nodeId: string, displayName: string) => void;
}

interface TreeNode extends OpcuaBrowseNode {
  children?: TreeNode[];
  isLoading?: boolean;
  isExpanded?: boolean;
}

const styles = {
  container: css`
    max-height: 400px;
    overflow: auto;
    padding: 8px;
  `,
  nodeRow: css`
    display: flex;
    align-items: center;
    padding: 4px 8px;
    cursor: pointer;
    border-radius: 4px;
    &:hover {
      background: rgba(255, 255, 255, 0.1);
    }
  `,
  nodeIcon: css`
    margin-left: 4px;
    margin-right: 8px;
  `,
  addIcon: css`
    margin-left: auto;
    cursor: pointer;
    opacity: 0.7;
    &:hover {
      opacity: 1;
    }
  `,
  spacer: css`
    width: 16px;
    display: inline-block;
  `,
};

/**
 * Node browser component for exploring OPC-UA server nodes
 *
 * Displays a tree view of the OPC-UA address space, allowing users to:
 * - Expand folders to view child nodes
 * - Select Variable nodes to add them to the query
 */
export const NodeBrowser: React.FC<Props> = ({ datasource, onSelect }) => {
  const [nodes, setNodes] = useState<TreeNode[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadRootNodes = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const rootNodes = await datasource.browseNodes();
      setNodes(rootNodes.map((n) => ({ ...n, isExpanded: false })));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to browse nodes');
    } finally {
      setLoading(false);
    }
  }, [datasource]);

  useEffect(() => {
    loadRootNodes();
  }, [loadRootNodes]);

  const loadChildren = async (nodeId: string): Promise<OpcuaBrowseNode[]> => {
    try {
      return await datasource.browseNodes(nodeId);
    } catch (err) {
      // Only log errors in development to avoid potential information leakage in production
      if (process.env.NODE_ENV === 'development') {
        console.error('Failed to load children:', err);
      }
      return [];
    }
  };

  const updateNodeAtPath = (nodes: TreeNode[], path: number[], updater: (node: TreeNode) => TreeNode): TreeNode[] => {
    if (path.length === 1) {
      const newNodes = [...nodes];
      newNodes[path[0]] = updater(newNodes[path[0]]);
      return newNodes;
    }

    const newNodes = [...nodes];
    newNodes[path[0]] = {
      ...newNodes[path[0]],
      children: updateNodeAtPath(newNodes[path[0]].children || [], path.slice(1), updater),
    };
    return newNodes;
  };

  const toggleNode = async (node: TreeNode, path: number[]) => {
    if (!node.hasChildren) {
      return;
    }

    if (node.isExpanded) {
      // Collapse
      setNodes((prev) => updateNodeAtPath(prev, path, (n) => ({ ...n, isExpanded: false })));
    } else {
      // Expand - show loading state
      setNodes((prev) =>
        updateNodeAtPath(prev, path, (n) => ({
          ...n,
          isExpanded: true,
          isLoading: true,
        }))
      );

      // Load children if not already loaded
      if (!node.children) {
        const children = await loadChildren(node.nodeId);
        setNodes((prev) =>
          updateNodeAtPath(prev, path, (n) => ({
            ...n,
            children: children.map((c) => ({ ...c, isExpanded: false })),
            isLoading: false,
          }))
        );
      } else {
        setNodes((prev) => updateNodeAtPath(prev, path, (n) => ({ ...n, isLoading: false })));
      }
    }
  };

  const renderNode = (node: TreeNode, path: number[], depth: number): React.ReactNode => {
    const isVariable = node.nodeClass === 'Variable';

    return (
      <div key={node.nodeId} style={{ marginLeft: depth * 20 }}>
        <div
          className={styles.nodeRow}
          onClick={() => toggleNode(node, path)}
          role="treeitem"
          aria-expanded={node.isExpanded}
          aria-label={node.displayName}
        >
          {node.hasChildren ? (
            node.isLoading ? (
              <Spinner size={12} />
            ) : (
              <Icon name={node.isExpanded ? 'angle-down' : 'angle-right'} />
            )
          ) : (
            <span className={styles.spacer} />
          )}
          <Icon name={isVariable ? 'tag-alt' : 'folder'} className={styles.nodeIcon} />
          <span>{node.displayName}</span>
          {isVariable && (
            <Icon
              name="plus-circle"
              className={styles.addIcon}
              onClick={(e) => {
                e.stopPropagation();
                onSelect(node.nodeId, node.displayName);
              }}
              title="Add to query"
            />
          )}
        </div>
        {node.isExpanded && node.children && (
          <div role="group">{node.children.map((child, index) => renderNode(child, [...path, index], depth + 1))}</div>
        )}
      </div>
    );
  };

  if (loading) {
    return (
      <div className={styles.container}>
        <Spinner />
        <span style={{ marginLeft: 8 }}>Loading nodes...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className={styles.container}>
        <Alert title="Error loading nodes" severity="error">
          {error}
        </Alert>
      </div>
    );
  }

  if (nodes.length === 0) {
    return (
      <div className={styles.container}>
        <Alert title="No nodes found" severity="info">
          The OPC-UA server returned no browseable nodes. Check that the data source is configured correctly.
        </Alert>
      </div>
    );
  }

  return (
    <div className={styles.container} role="tree">
      {nodes.map((node, index) => renderNode(node, [index], 0))}
    </div>
  );
};
