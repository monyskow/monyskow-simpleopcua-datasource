import React from 'react';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import {
  FieldSet,
  InlineField,
  Input,
  Select,
  SecretInput,
  SecretTextArea,
} from '@grafana/ui';
import {
  OpcuaDataSourceOptions,
  OpcuaSecureJsonData,
  AuthMethod,
  SecurityPolicy,
  SecurityMode,
} from '../../types';

type Props = DataSourcePluginOptionsEditorProps<
  OpcuaDataSourceOptions,
  OpcuaSecureJsonData
>;

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
export const ConfigEditor: React.FC<Props> = ({
  options,
  onOptionsChange,
}) => {
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onJsonDataChange = <K extends keyof OpcuaDataSourceOptions>(
    key: K,
    value: OpcuaDataSourceOptions[K]
  ) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        [key]: value,
      },
    });
  };

  const onSecureJsonDataChange = <K extends keyof OpcuaSecureJsonData>(
    key: K,
    value: string
  ) => {
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

        <InlineField
          label="Security Policy"
          labelWidth={20}
          tooltip="Encryption algorithm for the OPC-UA connection"
        >
          <Select
            width={30}
            options={SECURITY_POLICY_OPTIONS}
            value={jsonData.securityPolicy || 'None'}
            onChange={(v) => onJsonDataChange('securityPolicy', v.value!)}
            aria-label="Security Policy"
          />
        </InlineField>

        <InlineField
          label="Security Mode"
          labelWidth={20}
          tooltip="Security mode for message signing and encryption"
        >
          <Select
            width={30}
            options={SECURITY_MODE_OPTIONS}
            value={jsonData.securityMode || 'None'}
            onChange={(v) => onJsonDataChange('securityMode', v.value!)}
            aria-label="Security Mode"
          />
        </InlineField>

        <InlineField
          label="Timeout (seconds)"
          labelWidth={20}
          tooltip="Connection timeout in seconds"
        >
          <Input
            type="number"
            width={10}
            value={jsonData.timeout || 10}
            onChange={(e) =>
              onJsonDataChange('timeout', parseInt(e.currentTarget.value, 10) || 10)
            }
            aria-label="Timeout"
          />
        </InlineField>
      </FieldSet>

      <FieldSet label="Authentication">
        <InlineField
          label="Method"
          labelWidth={20}
          tooltip="Authentication method to use when connecting to the OPC-UA server"
        >
          <Select
            width={30}
            options={AUTH_METHOD_OPTIONS}
            value={jsonData.authMethod || 'anonymous'}
            onChange={(v) => onJsonDataChange('authMethod', v.value!)}
            aria-label="Authentication Method"
          />
        </InlineField>

        {jsonData.authMethod === 'userpass' && (
          <>
            <InlineField label="Username" labelWidth={20}>
              <SecretInput
                width={30}
                isConfigured={!!secureJsonFields?.username}
                value={secureJsonData?.username || ''}
                onChange={(e) =>
                  onSecureJsonDataChange('username', e.currentTarget.value)
                }
                onReset={() => onResetSecureJsonData('username')}
                aria-label="Username"
              />
            </InlineField>
            <InlineField label="Password" labelWidth={20}>
              <SecretInput
                width={30}
                isConfigured={!!secureJsonFields?.password}
                value={secureJsonData?.password || ''}
                onChange={(e) =>
                  onSecureJsonDataChange('password', e.currentTarget.value)
                }
                onReset={() => onResetSecureJsonData('password')}
                aria-label="Password"
              />
            </InlineField>
          </>
        )}

        {jsonData.authMethod === 'certificate' && (
          <>
            <InlineField
              label="Certificate (PEM)"
              labelWidth={20}
              tooltip="PEM-encoded X.509 certificate"
            >
              <SecretTextArea
                cols={50}
                rows={5}
                isConfigured={!!secureJsonFields?.certificate}
                value={secureJsonData?.certificate || ''}
                onChange={(e) =>
                  onSecureJsonDataChange('certificate', e.currentTarget.value)
                }
                onReset={() => onResetSecureJsonData('certificate')}
                placeholder="-----BEGIN CERTIFICATE-----"
                aria-label="Certificate"
              />
            </InlineField>
            <InlineField
              label="Private Key (PEM)"
              labelWidth={20}
              tooltip="PEM-encoded private key"
            >
              <SecretTextArea
                cols={50}
                rows={5}
                isConfigured={!!secureJsonFields?.privateKey}
                value={secureJsonData?.privateKey || ''}
                onChange={(e) =>
                  onSecureJsonDataChange('privateKey', e.currentTarget.value)
                }
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
