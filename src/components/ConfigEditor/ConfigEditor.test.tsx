/* eslint-disable @typescript-eslint/no-deprecated */
import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { of, Observable } from 'rxjs';
import { ConfigEditor } from './ConfigEditor';
import { OpcuaDataSourceOptions, OpcuaSecureJsonData } from '../../types';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';

const mockFetch = jest.fn();

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  getBackendSrv: () => ({
    fetch: mockFetch,
  }),
  getTemplateSrv: () => ({
    replace: (s: string) => s,
  }),
}));

// Combobox from @grafana/ui uses canvas.measureText which isn't available in jsdom.
// Replace it with a plain <select> that calls onChange the same way the real Combobox does.
jest.mock('@grafana/ui', () => {
  const actual = jest.requireActual('@grafana/ui');
  const MockCombobox = ({
    options,
    value,
    onChange,
    ...rest
  }: {
    options: Array<{ label: string; value: string }>;
    value: string;
    onChange: (opt: { value: string }) => void;
    [key: string]: unknown;
  }) => (
    <select
      value={value}
      onChange={(e) => onChange({ value: e.target.value })}
      data-testid={rest['aria-label'] as string}
    >
      {options.map((o) => (
        <option key={o.value} value={o.value}>
          {o.label}
        </option>
      ))}
    </select>
  );
  return { ...actual, Combobox: MockCombobox };
});

type Options = DataSourcePluginOptionsEditorProps<OpcuaDataSourceOptions, OpcuaSecureJsonData>;

function makeOptions(overrides: Partial<Options['options']> = {}): Options['options'] {
  return {
    id: 0,
    uid: '',
    orgId: 1,
    name: 'Test OPC-UA',
    type: 'monyskow-simpleopcua-datasource',
    typeName: 'Simple OPC-UA',
    typeLogoUrl: '',
    access: 'proxy',
    url: '',
    user: '',
    database: '',
    basicAuth: false,
    basicAuthUser: '',
    isDefault: false,
    readOnly: false,
    withCredentials: false,
    secureJsonFields: {},
    jsonData: {
      endpoint: 'opc.tcp://localhost:4840',
      securityPolicy: 'None',
      securityMode: 'None',
      authMethod: 'anonymous',
      timeout: 10,
    },
    secureJsonData: {},
    ...overrides,
  };
}

function renderEditor(optionOverrides: Partial<Options['options']> = {}, onOptionsChange = jest.fn()) {
  const options = makeOptions(optionOverrides);
  return {
    onOptionsChange,
    ...render(<ConfigEditor options={options} onOptionsChange={onOptionsChange} />),
  };
}

describe('ConfigEditor', () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  describe('initial render', () => {
    it('shows the Connection fieldset', () => {
      renderEditor();
      expect(screen.getByText('Connection')).toBeInTheDocument();
    });

    it('shows the Authentication fieldset', () => {
      renderEditor();
      expect(screen.getByText('Authentication')).toBeInTheDocument();
    });

    it('shows the Endpoint URL input with default value', () => {
      renderEditor();
      const input = screen.getByLabelText('Endpoint URL');
      expect(input).toBeInTheDocument();
      expect(input).toHaveValue('opc.tcp://localhost:4840');
    });

    it('does not show Client Certificate section when security mode is None', () => {
      renderEditor();
      expect(screen.queryByText('Client Certificate')).not.toBeInTheDocument();
    });

    it('does not show username/password fields for anonymous auth', () => {
      renderEditor();
      expect(screen.queryByLabelText('Username')).not.toBeInTheDocument();
      expect(screen.queryByLabelText('Password')).not.toBeInTheDocument();
    });
  });

  describe('URL input interaction', () => {
    it('calls onOptionsChange with updated endpoint when URL changes', () => {
      const onOptionsChange = jest.fn();
      renderEditor({}, onOptionsChange);

      const input = screen.getByLabelText('Endpoint URL');
      fireEvent.change(input, { target: { value: 'opc.tcp://myserver:4840' } });

      expect(onOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          jsonData: expect.objectContaining({
            endpoint: 'opc.tcp://myserver:4840',
          }),
        })
      );
    });
  });

  describe('security mode — Client Certificate section', () => {
    it('shows Client Certificate section when securityMode is Sign', () => {
      renderEditor({
        jsonData: {
          endpoint: 'opc.tcp://localhost:4840',
          securityPolicy: 'None',
          securityMode: 'Sign',
          authMethod: 'anonymous',
          timeout: 10,
        },
      });
      expect(screen.getByText('Client Certificate')).toBeInTheDocument();
    });

    it('shows Client Certificate section when securityMode is SignAndEncrypt', () => {
      renderEditor({
        jsonData: {
          endpoint: 'opc.tcp://localhost:4840',
          securityPolicy: 'None',
          securityMode: 'SignAndEncrypt',
          authMethod: 'anonymous',
          timeout: 10,
        },
      });
      expect(screen.getByText('Client Certificate')).toBeInTheDocument();
    });

    it('does not show Client Certificate section when authMethod is certificate (user provides cert)', () => {
      renderEditor({
        jsonData: {
          endpoint: 'opc.tcp://localhost:4840',
          securityPolicy: 'None',
          securityMode: 'Sign',
          authMethod: 'certificate',
          timeout: 10,
        },
      });
      expect(screen.queryByText('Client Certificate')).not.toBeInTheDocument();
    });

    it('shows Generate Certificate button when no cert is configured', () => {
      renderEditor({
        jsonData: {
          endpoint: 'opc.tcp://localhost:4840',
          securityPolicy: 'None',
          securityMode: 'Sign',
          authMethod: 'anonymous',
          timeout: 10,
        },
      });
      expect(screen.getByRole('button', { name: 'Generate Certificate' })).toBeInTheDocument();
    });

    it('shows "no certificate" status when no cert configured', () => {
      renderEditor({
        jsonData: {
          endpoint: 'opc.tcp://localhost:4840',
          securityPolicy: 'None',
          securityMode: 'Sign',
          authMethod: 'anonymous',
          timeout: 10,
        },
      });
      expect(screen.getByText(/No certificate configured/)).toBeInTheDocument();
    });

    it('shows saved status when secureJsonFields has clientCert and clientKey', () => {
      renderEditor({
        jsonData: {
          endpoint: 'opc.tcp://localhost:4840',
          securityPolicy: 'None',
          securityMode: 'Sign',
          authMethod: 'anonymous',
          timeout: 10,
        },
        secureJsonFields: { clientCert: true, clientKey: true },
      });
      expect(screen.getByText(/Certificate configured \(saved\)/)).toBeInTheDocument();
    });

    it('shows pending status when cert is in secureJsonData but not secureJsonFields', () => {
      renderEditor({
        jsonData: {
          endpoint: 'opc.tcp://localhost:4840',
          securityPolicy: 'None',
          securityMode: 'Sign',
          authMethod: 'anonymous',
          timeout: 10,
        },
        secureJsonFields: {},
        secureJsonData: { clientCert: 'CERT', clientKey: 'KEY' },
      });
      expect(screen.getByText(/Certificate generated - click Save/)).toBeInTheDocument();
    });
  });

  describe('Generate Certificate — saved datasource', () => {
    it('calls fetch and updates secureJsonData on success', async () => {
      mockFetch.mockReturnValue(of({ data: { clientCert: '---CERT---', clientKey: '---KEY---' } }));

      const onOptionsChange = jest.fn();
      renderEditor(
        {
          id: 42,
          jsonData: {
            endpoint: 'opc.tcp://localhost:4840',
            securityPolicy: 'None',
            securityMode: 'Sign',
            authMethod: 'anonymous',
            timeout: 10,
          },
        },
        onOptionsChange
      );

      const btn = screen.getByRole('button', { name: 'Generate Certificate' });
      await act(async () => {
        fireEvent.click(btn);
      });

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalledWith(
          expect.objectContaining({
            url: '/api/datasources/42/resources/generate-certificate',
            method: 'GET',
          })
        );
      });

      expect(onOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          secureJsonData: expect.objectContaining({
            clientCert: '---CERT---',
            clientKey: '---KEY---',
          }),
        })
      );
    });

    it('shows error message on fetch failure', async () => {
      mockFetch.mockReturnValue(new Observable((sub) => sub.error(new Error('Network error'))));

      renderEditor({
        id: 42,
        jsonData: {
          endpoint: 'opc.tcp://localhost:4840',
          securityPolicy: 'None',
          securityMode: 'Sign',
          authMethod: 'anonymous',
          timeout: 10,
        },
      });

      const btn = screen.getByRole('button', { name: 'Generate Certificate' });
      await act(async () => {
        fireEvent.click(btn);
      });

      await waitFor(() => {
        expect(screen.getByText('Network error')).toBeInTheDocument();
      });
    });
  });

  describe('AbortController cleanup on unmount', () => {
    it('does not call onOptionsChange after unmount during pending fetch', async () => {
      // The real discrimination path: when abortSignal fires, the Observable errors,
      // firstValueFrom rejects, the catch block runs, and onOptionsChange is never reached.
      // If the useEffect cleanup (abortControllerRef.current?.abort()) is removed,
      // the emit resolves the Observable successfully and onOptionsChange IS called.
      let emit: (v: unknown) => void = () => {};
      mockFetch.mockImplementation(
        (config: { abortSignal?: AbortSignal; [key: string]: unknown }) =>
          new Observable((sub) => {
            // Honor the abortSignal — this is what the real fetch does
            config.abortSignal?.addEventListener('abort', () => {
              sub.error(new DOMException('The user aborted a request.', 'AbortError'));
            });
            emit = (v) => {
              sub.next(v);
              sub.complete();
            };
          })
      );

      const onOptionsChange = jest.fn();

      const { unmount } = renderEditor(
        {
          id: 42,
          jsonData: {
            endpoint: 'opc.tcp://localhost:4840',
            securityPolicy: 'None',
            securityMode: 'Sign',
            authMethod: 'anonymous',
            timeout: 10,
          },
        },
        onOptionsChange
      );

      const btn = screen.getByRole('button', { name: 'Generate Certificate' });
      fireEvent.click(btn);

      // Give the click handler a tick to start the fetch and wire up the signal listener
      await Promise.resolve();

      // Unmount BEFORE the fetch resolves — this fires abort via useEffect cleanup
      unmount();

      // NOW resolve the fetch — the abort should have already errored the Observable,
      // so this emit arrives after the subscription is dead
      emit({ data: { clientCert: 'CERT', clientKey: 'KEY' } });

      // Drain the event loop
      await act(async () => {});

      // onOptionsChange must NOT have been called with cert payload
      const callsWithCert = onOptionsChange.mock.calls.filter((args) => args[0]?.secureJsonData?.clientCert);
      expect(callsWithCert).toHaveLength(0);
    });
  });

  describe('Combobox interactions', () => {
    it('calls onOptionsChange with updated securityMode when Security Mode combobox changes', () => {
      const onOptionsChange = jest.fn();
      renderEditor({}, onOptionsChange);

      // Combobox order: 0=Security Policy, 1=Security Mode, 2=Auth Method
      const selects = screen.getAllByRole('combobox');
      fireEvent.change(selects[1], { target: { value: 'Sign' } });

      expect(onOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          jsonData: expect.objectContaining({ securityMode: 'Sign' }),
        })
      );
    });

    it('calls onOptionsChange with updated authMethod when Auth Method combobox changes', () => {
      const onOptionsChange = jest.fn();
      renderEditor({}, onOptionsChange);

      // Combobox order: 0=Security Policy, 1=Security Mode, 2=Auth Method
      const selects = screen.getAllByRole('combobox');
      fireEvent.change(selects[2], { target: { value: 'userpass' } });

      expect(onOptionsChange).toHaveBeenCalledWith(
        expect.objectContaining({
          jsonData: expect.objectContaining({ authMethod: 'userpass' }),
        })
      );
    });
  });

  describe('Authentication — userpass fields', () => {
    it('shows username and password inputs when authMethod is userpass', () => {
      renderEditor({
        jsonData: {
          endpoint: 'opc.tcp://localhost:4840',
          securityPolicy: 'None',
          securityMode: 'None',
          authMethod: 'userpass',
          timeout: 10,
        },
      });
      expect(screen.getByLabelText('Username')).toBeInTheDocument();
      expect(screen.getByLabelText('Password')).toBeInTheDocument();
    });

    it('shows certificate fields when authMethod is certificate', () => {
      renderEditor({
        jsonData: {
          endpoint: 'opc.tcp://localhost:4840',
          securityPolicy: 'None',
          securityMode: 'None',
          authMethod: 'certificate',
          timeout: 10,
        },
      });
      expect(screen.getByLabelText('Certificate')).toBeInTheDocument();
      expect(screen.getByLabelText('Private Key')).toBeInTheDocument();
    });

    it('does not show username/password for certificate auth', () => {
      renderEditor({
        jsonData: {
          endpoint: 'opc.tcp://localhost:4840',
          securityPolicy: 'None',
          securityMode: 'None',
          authMethod: 'certificate',
          timeout: 10,
        },
      });
      expect(screen.queryByLabelText('Username')).not.toBeInTheDocument();
      expect(screen.queryByLabelText('Password')).not.toBeInTheDocument();
    });
  });
});
