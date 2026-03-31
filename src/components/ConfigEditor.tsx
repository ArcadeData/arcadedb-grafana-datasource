import React from 'react';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { ArcadeDBDataSourceOptions, ArcadeDBSecureJsonData } from '../types';

type Props = DataSourcePluginOptionsEditorProps<ArcadeDBDataSourceOptions, ArcadeDBSecureJsonData>;

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onDatabaseChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: { ...jsonData, database: event.target.value },
    });
  };

  const onPasswordChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: { ...secureJsonData, password: event.target.value },
    });
  };

  const onPasswordReset = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: { ...secureJsonFields, password: false },
      secureJsonData: { ...secureJsonData, password: '' },
    });
  };

  return (
    <>
      <h3 className="page-heading">ArcadeDB Connection</h3>

      <InlineField label="Database" labelWidth={20} tooltip="ArcadeDB database name">
        <Input
          width={40}
          value={jsonData.database || ''}
          onChange={onDatabaseChange}
          placeholder="mydb"
        />
      </InlineField>

      <InlineField label="Password" labelWidth={20} tooltip="ArcadeDB password (username is set in the HTTP Auth section above)">
        <SecretInput
          width={40}
          isConfigured={secureJsonFields?.password}
          value={secureJsonData?.password || ''}
          onChange={onPasswordChange}
          onReset={onPasswordReset}
          placeholder="password"
        />
      </InlineField>
    </>
  );
}
