import React from 'react';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { ArcadeDBDataSourceOptions, ArcadeDBSecureJsonData } from '../types';

type Props = DataSourcePluginOptionsEditorProps<ArcadeDBDataSourceOptions, ArcadeDBSecureJsonData>;

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onUrlChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({ ...options, url: event.target.value });
  };

  const onUsernameChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({ ...options, basicAuthUser: event.target.value, basicAuth: true });
  };

  const onDatabaseChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: { ...jsonData, database: event.target.value },
    });
  };

  const onPasswordChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      basicAuth: true,
      secureJsonData: { ...secureJsonData, basicAuthPassword: event.target.value, password: event.target.value },
    });
  };

  const onPasswordReset = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: { ...secureJsonFields, basicAuthPassword: false, password: false },
      secureJsonData: { ...secureJsonData, basicAuthPassword: '', password: '' },
    });
  };

  return (
    <>
      <h3 className="page-heading">ArcadeDB Connection</h3>

      <InlineField label="URL" labelWidth={20} tooltip="ArcadeDB server HTTP URL (e.g. http://localhost:2480)">
        <Input
          width={40}
          value={options.url || ''}
          onChange={onUrlChange}
          placeholder="http://localhost:2480"
        />
      </InlineField>

      <InlineField label="Database" labelWidth={20} tooltip="ArcadeDB database name">
        <Input
          width={40}
          value={jsonData.database || ''}
          onChange={onDatabaseChange}
          placeholder="mydb"
        />
      </InlineField>

      <InlineField label="Username" labelWidth={20} tooltip="ArcadeDB username">
        <Input
          width={40}
          value={options.basicAuthUser || ''}
          onChange={onUsernameChange}
          placeholder="root"
        />
      </InlineField>

      <InlineField label="Password" labelWidth={20} tooltip="ArcadeDB password">
        <SecretInput
          width={40}
          isConfigured={secureJsonFields?.basicAuthPassword || secureJsonFields?.password}
          value={secureJsonData?.password || ''}
          onChange={onPasswordChange}
          onReset={onPasswordReset}
          placeholder="password"
        />
      </InlineField>
    </>
  );
}
