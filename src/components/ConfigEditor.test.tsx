import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { ConfigEditor } from './ConfigEditor';
import { DataSourcePluginOptionsEditorProps, DataSourceSettings } from '@grafana/data';
import { ArcadeDBDataSourceOptions, ArcadeDBSecureJsonData } from '../types';

jest.mock('@grafana/ui', () => ({
  InlineField: ({ label, children }: any) => <label>{label}{children}</label>,
  Input: ({ value, onChange, placeholder, ...rest }: any) => (
    <input value={value} onChange={onChange} placeholder={placeholder} aria-label={placeholder} {...rest} />
  ),
  SecretInput: ({ isConfigured, value, onChange, onReset, placeholder }: any) => (
    isConfigured ? (
      <div><span>configured</span><button onClick={onReset}>Reset</button></div>
    ) : (
      <input value={value} onChange={onChange} placeholder={placeholder} aria-label={placeholder} />
    )
  ),
}));

type Props = DataSourcePluginOptionsEditorProps<ArcadeDBDataSourceOptions, ArcadeDBSecureJsonData>;

function makeProps(overrides?: Partial<DataSourceSettings<ArcadeDBDataSourceOptions, ArcadeDBSecureJsonData>>): Props {
  const defaults: DataSourceSettings<ArcadeDBDataSourceOptions, ArcadeDBSecureJsonData> = {
    id: 1,
    uid: 'test-uid',
    orgId: 1,
    name: 'ArcadeDB',
    type: 'arcade-arcadedb-datasource',
    typeName: 'ArcadeDB',
    typeLogoUrl: '',
    access: 'proxy',
    url: '',
    basicAuth: false,
    basicAuthUser: '',
    isDefault: false,
    withCredentials: false,
    jsonData: { database: '' },
    secureJsonFields: {},
    secureJsonData: { password: '', basicAuthPassword: '' },
    version: 1,
    readOnly: false,
    database: '',
    user: '',
    apiVersion: '',
  };

  return {
    options: { ...defaults, ...overrides },
    onOptionsChange: jest.fn(),
  } as unknown as Props;
}

describe('ConfigEditor', () => {
  it('renders all fields', () => {
    const props = makeProps();
    render(<ConfigEditor {...props} />);

    expect(screen.getByPlaceholderText('http://localhost:2480')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('mydb')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('root')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('password')).toBeInTheDocument();
  });

  it('calls onChange when URL is updated', () => {
    const props = makeProps();
    render(<ConfigEditor {...props} />);

    fireEvent.change(screen.getByPlaceholderText('http://localhost:2480'), {
      target: { value: 'http://myhost:2480' },
    });

    expect(props.onOptionsChange).toHaveBeenCalledWith(
      expect.objectContaining({ url: 'http://myhost:2480' })
    );
  });

  it('calls onChange when database is updated', () => {
    const props = makeProps();
    render(<ConfigEditor {...props} />);

    fireEvent.change(screen.getByPlaceholderText('mydb'), {
      target: { value: 'testdb' },
    });

    expect(props.onOptionsChange).toHaveBeenCalledWith(
      expect.objectContaining({
        jsonData: expect.objectContaining({ database: 'testdb' }),
      })
    );
  });

  it('calls onChange when username is updated', () => {
    const props = makeProps();
    render(<ConfigEditor {...props} />);

    fireEvent.change(screen.getByPlaceholderText('root'), {
      target: { value: 'admin' },
    });

    expect(props.onOptionsChange).toHaveBeenCalledWith(
      expect.objectContaining({ basicAuthUser: 'admin', basicAuth: true })
    );
  });

  it('handles password reset', () => {
    const props = makeProps({
      secureJsonFields: { basicAuthPassword: true, password: true },
    });
    render(<ConfigEditor {...props} />);

    expect(screen.getByText('configured')).toBeInTheDocument();

    fireEvent.click(screen.getByText('Reset'));

    expect(props.onOptionsChange).toHaveBeenCalledWith(
      expect.objectContaining({
        secureJsonFields: expect.objectContaining({
          basicAuthPassword: false,
          password: false,
        }),
        secureJsonData: expect.objectContaining({
          basicAuthPassword: '',
          password: '',
        }),
      })
    );
  });

  it('populates fields from existing options', () => {
    const props = makeProps({
      url: 'http://arcade:2480',
      basicAuthUser: 'myuser',
      jsonData: { database: 'production' },
      secureJsonData: { password: 'secret', basicAuthPassword: 'secret' },
      secureJsonFields: {},
    });
    render(<ConfigEditor {...props} />);

    expect(screen.getByPlaceholderText('http://localhost:2480')).toHaveValue('http://arcade:2480');
    expect(screen.getByPlaceholderText('mydb')).toHaveValue('production');
    expect(screen.getByPlaceholderText('root')).toHaveValue('myuser');
    expect(screen.getByPlaceholderText('password')).toHaveValue('secret');
  });
});
