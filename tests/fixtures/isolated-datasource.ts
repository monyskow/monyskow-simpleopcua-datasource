import { test as base } from '@playwright/test';
import { randomUUID } from 'crypto';

export interface IsolatedDatasource {
  id: number;
  uid: string;
  name: string;
}

/**
 * Fixture that creates a fresh, per-test OPC-UA datasource via the Grafana HTTP API
 * and deletes it after the test completes (regardless of outcome).
 *
 * Motivation: multiple workers concurrently PATCHing a shared datasource triggers
 * Grafana's optimistic concurrency control ("Datasource has already been updated by
 * someone else", HTTP 409). Per-test isolation eliminates the shared mutable state.
 *
 * The UID is composed of workerIndex + random suffix so parallel workers never
 * collide even when both enter beforeEach in the same millisecond.
 */
export const test = base.extend<{ isolatedDatasource: IsolatedDatasource }>({
  isolatedDatasource: async ({ request }, use, testInfo) => {
    const uid = `opcua-iso-${testInfo.workerIndex}-${randomUUID().slice(0, 8)}`;
    const name = `OPC-UA Isolated ${testInfo.workerIndex}-${randomUUID().slice(0, 8)}`;

    const response = await request.post('/api/datasources', {
      data: {
        name,
        uid,
        type: 'monyskow-simpleopcua-datasource',
        access: 'proxy',
        // isDefault omitted — Grafana enforces a single default per org;
        // provisioned "OPC-UA Test Server" already holds that slot.
        jsonData: {
          endpoint: 'opc.tcp://opcua-server:4840',
          securityPolicy: 'None',
          securityMode: 'None',
          authMethod: 'anonymous',
          timeout: 10,
        },
        editable: true,
      },
    });

    if (!response.ok()) {
      const body = await response.text();
      throw new Error(`Failed to create isolated datasource (HTTP ${response.status()}): ${body}`);
    }

    const created: { datasource: IsolatedDatasource } = await response.json();
    const ds = created.datasource;

    await use(ds);

    // Teardown: delete regardless of test outcome; ignore 404 in case the test
    // itself already deleted the resource.
    try {
      await request.delete(`/api/datasources/uid/${ds.uid}`);
    } catch {
      // Deletion errors are non-fatal — log and move on.
      console.warn(`[isolated-datasource] Failed to delete DS uid=${ds.uid} (may already be gone)`);
    }
  },
});

export { expect } from '@playwright/test';
