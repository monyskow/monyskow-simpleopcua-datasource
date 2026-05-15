import React, { useEffect, useRef, useState } from 'react';
import { firstValueFrom } from 'rxjs';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { getBackendSrv } from '@grafana/runtime';
import { FieldSet, InlineField, Input, Combobox, SecretInput, SecretTextArea, Button, Alert } from '@grafana/ui';
import { OpcuaDataSourceOptions, OpcuaSecureJsonData, AuthMethod, SecurityPolicy, SecurityMode } from '../../types';

type Props = DataSourcePluginOptionsEditorProps<OpcuaDataSourceOptions, OpcuaSecureJsonData>;

const AUTH_METHOD_OPTIONS = [
  { label: 'Anonymous', value: 'anonymous' as AuthMethod },
  { label: 'Username / Password', value: 'userpass' as AuthMethod },
  { label: 'Certificate', value: 'certificate' as AuthMethod },
];

const SECURITY_POLICY_OPTIONS = [
  { label: 'None', value: 'None' as SecurityPolicy },
  { label: 'Basic128Rsa15', value: 'Basic128Rsa15' as SecurityPolicy },
  { label: 'Basic256', value: 'Basic256' as SecurityPolicy },
  { label: 'Basic256Sha256', value: 'Basic256Sha256' as SecurityPolicy },
  {
    label: 'Aes128_Sha256_RsaOaep',
    value: 'Aes128_Sha256_RsaOaep' as SecurityPolicy,
  },
  {
    label: 'Aes256_Sha256_RsaPss',
    value: 'Aes256_Sha256_RsaPss' as SecurityPolicy,
  },
];

const SECURITY_MODE_OPTIONS = [
  { label: 'None', value: 'None' as SecurityMode },
  { label: 'Sign', value: 'Sign' as SecurityMode },
  { label: 'Sign and Encrypt', value: 'SignAndEncrypt' as SecurityMode },
];

/**
 * Configuration editor for the OPC-UA data source
 *
 * Handles:
 * - Connection settings (endpoint, security policy, security mode, timeout)
 * - Authentication method selection and credential input
 * - Secure storage of credentials via secureJsonData
 */
export const ConfigEditor: React.FC<Props> = ({ options, onOptionsChange }) => {
  const { jsonData, secureJsonFields, secureJsonData } = options;
  const [isGenerating, setIsGenerating] = useState(false);
  const [generateError, setGenerateError] = useState<string | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  useEffect(() => {
    return () => {
      abortControllerRef.current?.abort();
    };
  }, []);

  // Check if security mode requires a client certificate
  const needsClientCertificate =
    (jsonData.securityMode === 'Sign' || jsonData.securityMode === 'SignAndEncrypt') &&
    jsonData.authMethod !== 'certificate'; // Certificate auth uses user-provided cert

  // Check if client certificate is already configured (saved to backend)
  const hasClientCertificateSaved = !!secureJsonFields?.clientCert && !!secureJsonFields?.clientKey;

  // Check if client certificate is pending (generated but not yet saved)
  const hasClientCertificatePending =
    !hasClientCertificateSaved && !!secureJsonData?.clientCert && !!secureJsonData?.clientKey;

  // Either saved or pending
  const hasClientCertificate = hasClientCertificateSaved || hasClientCertificatePending;

  const onJsonDataChange = <K extends keyof OpcuaDataSourceOptions>(key: K, value: OpcuaDataSourceOptions[K]) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        [key]: value,
      },
    });
  };

  const onSecureJsonDataChange = <K extends keyof OpcuaSecureJsonData>(key: K, value: string) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...secureJsonData,
        [key]: value,
      },
    });
  };

  const onResetSecureJsonData = (key: keyof OpcuaSecureJsonData) => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...secureJsonFields,
        [key]: false,
      },
      secureJsonData: {
        ...secureJsonData,
        [key]: '',
      },
    });
  };

  const onGenerateCertificate = async () => {
    setIsGenerating(true);
    setGenerateError(null);

    try {
      // Call the backend to generate a certificate
      // Note: For unsaved datasources, we need a different approach
      if (!options.id) {
        // Datasource not saved yet - we can't call the backend resource API
        // Instead, generate a temporary message
        setGenerateError('Please save the datasource first, then click Generate Certificate.');
        setIsGenerating(false);
        return;
      }

      abortControllerRef.current?.abort();
      abortControllerRef.current = new AbortController();
      const response = await firstValueFrom(
        getBackendSrv().fetch<{ clientCert: string; clientKey: string }>({
          url: `/api/datasources/${options.id}/resources/generate-certificate`,
          method: 'GET',
          abortSignal: abortControllerRef.current.signal,
        })
      );

      if (response?.data) {
        // Store the certificate in secureJsonData
        onOptionsChange({
          ...options,
          secureJsonData: {
            ...secureJsonData,
            clientCert: response.data.clientCert,
            clientKey: response.data.clientKey,
          },
        });
      }
    } catch (error) {
      setGenerateError(error instanceof Error ? error.message : 'Failed to generate certificate');
    } finally {
      setIsGenerating(false);
    }
  };

  const onResetClientCertificate = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...secureJsonFields,
        clientCert: false,
        clientKey: false,
      },
      secureJsonData: {
        ...secureJsonData,
        clientCert: '',
        clientKey: '',
      },
    });
  };

  return (
    <>
      <FieldSet label="Connection">
        <InlineField
          label="Endpoint URL"
          labelWidth={20}
          tooltip="OPC-UA server endpoint (e.g., opc.tcp://localhost:4840)"
        >
          <Input
            width={50}
            value={jsonData.endpoint || ''}
            onChange={(e) => onJsonDataChange('endpoint', e.currentTarget.value)}
            placeholder="opc.tcp://localhost:4840"
            aria-label="Endpoint URL"
          />
        </InlineField>

        <InlineField label="Security Policy" labelWidth={20} tooltip="Encryption algorithm for the OPC-UA connection">
          <Combobox
            width={30}
            options={SECURITY_POLICY_OPTIONS}
            value={jsonData.securityPolicy || 'None'}
            onChange={(v) => onJsonDataChange('securityPolicy', v.value)}
          />
        </InlineField>

        <InlineField label="Security Mode" labelWidth={20} tooltip="Security mode for message signing and encryption">
          <Combobox
            width={30}
            options={SECURITY_MODE_OPTIONS}
            value={jsonData.securityMode || 'None'}
            onChange={(v) => onJsonDataChange('securityMode', v.value)}
          />
        </InlineField>

        <InlineField label="Timeout (seconds)" labelWidth={20} tooltip="Connection timeout in seconds">
          <Input
            type="number"
            width={10}
            value={jsonData.timeout || 10}
            onChange={(e) => onJsonDataChange('timeout', parseInt(e.currentTarget.value, 10) || 10)}
            aria-label="Timeout"
          />
        </InlineField>
      </FieldSet>

      {needsClientCertificate && (
        <FieldSet label="Client Certificate">
          <p style={{ marginBottom: '16px', color: '#8e8e8e' }}>
            Secure connections (Sign/SignAndEncrypt) require a client certificate. Generate one and save the datasource
            to persist it across Grafana restarts.
          </p>

          {generateError && (
            <Alert title="Certificate Generation Error" severity="error" style={{ marginBottom: '16px' }}>
              {generateError}
            </Alert>
          )}

          <InlineField label="Status" labelWidth={20}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', height: '32px' }}>
              {hasClientCertificateSaved ? (
                <>
                  <span style={{ color: '#73bf69' }}>✓ Certificate configured (saved)</span>
                  <Button variant="secondary" size="sm" onClick={onResetClientCertificate}>
                    Reset
                  </Button>
                </>
              ) : hasClientCertificatePending ? (
                <>
                  <span style={{ color: '#ff9830' }}>⚠ Certificate generated - click Save &amp; Test to persist</span>
                  <Button variant="secondary" size="sm" onClick={onResetClientCertificate}>
                    Reset
                  </Button>
                </>
              ) : (
                <span style={{ color: '#f2495c' }}>✗ No certificate configured</span>
              )}
            </div>
          </InlineField>

          {!hasClientCertificate && (
            <InlineField label="" labelWidth={20}>
              <Button onClick={onGenerateCertificate} disabled={isGenerating}>
                {isGenerating ? 'Generating...' : 'Generate Certificate'}
              </Button>
            </InlineField>
          )}

          {hasClientCertificate && (
            <p style={{ marginTop: '8px', color: '#8e8e8e', fontSize: '12px' }}>
              Remember to add this certificate to your OPC-UA server&apos;s trusted certificates list.
            </p>
          )}
        </FieldSet>
      )}

      <FieldSet label="Authentication">
        <InlineField
          label="Method"
          labelWidth={20}
          tooltip="Authentication method to use when connecting to the OPC-UA server"
        >
          <Combobox
            width={30}
            options={AUTH_METHOD_OPTIONS}
            value={jsonData.authMethod || 'anonymous'}
            onChange={(v) => onJsonDataChange('authMethod', v.value)}
          />
        </InlineField>

        {jsonData.authMethod === 'userpass' && (
          <>
            <InlineField label="Username" labelWidth={20}>
              <SecretInput
                width={30}
                isConfigured={!!secureJsonFields?.username}
                value={secureJsonData?.username || ''}
                onChange={(e) => onSecureJsonDataChange('username', e.currentTarget.value)}
                onReset={() => onResetSecureJsonData('username')}
                aria-label="Username"
              />
            </InlineField>
            <InlineField label="Password" labelWidth={20}>
              <SecretInput
                width={30}
                isConfigured={!!secureJsonFields?.password}
                value={secureJsonData?.password || ''}
                onChange={(e) => onSecureJsonDataChange('password', e.currentTarget.value)}
                onReset={() => onResetSecureJsonData('password')}
                aria-label="Password"
              />
            </InlineField>
          </>
        )}

        {jsonData.authMethod === 'certificate' && (
          <>
            <InlineField label="Certificate (PEM)" labelWidth={20} tooltip="PEM-encoded X.509 certificate">
              <SecretTextArea
                cols={50}
                rows={5}
                isConfigured={!!secureJsonFields?.certificate}
                value={secureJsonData?.certificate || ''}
                onChange={(e) => onSecureJsonDataChange('certificate', e.currentTarget.value)}
                onReset={() => onResetSecureJsonData('certificate')}
                placeholder="-----BEGIN CERTIFICATE-----"
                aria-label="Certificate"
              />
            </InlineField>
            <InlineField label="Private Key (PEM)" labelWidth={20} tooltip="PEM-encoded private key">
              <SecretTextArea
                cols={50}
                rows={5}
                isConfigured={!!secureJsonFields?.privateKey}
                value={secureJsonData?.privateKey || ''}
                onChange={(e) => onSecureJsonDataChange('privateKey', e.currentTarget.value)}
                onReset={() => onResetSecureJsonData('privateKey')}
                placeholder="-----BEGIN RSA PRIVATE KEY-----"
                aria-label="Private Key"
              />
            </InlineField>
          </>
        )}
      </FieldSet>
    </>
  );
};
