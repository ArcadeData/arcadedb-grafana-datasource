import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { TimeSeriesEditor } from './TimeSeriesEditor';
import { QueryMode, ArcadeDBQuery, ArcadeDBMetadata } from '../types';

const mockMetadata: ArcadeDBMetadata = {
  types: [
    {
      name: 'weather',
      fields: [
        { name: 'temperature', dataType: 'DOUBLE' },
        { name: 'humidity', dataType: 'DOUBLE' },
      ],
      tags: [{ name: 'location', dataType: 'STRING' }],
    },
    {
      name: 'cpu',
      fields: [{ name: 'usage', dataType: 'DOUBLE' }],
      tags: [],
    },
  ],
  aggregationTypes: ['AVG', 'SUM', 'MIN', 'MAX', 'COUNT'],
};

jest.mock('@grafana/ui', () => ({
  InlineField: ({ label, children }: any) => (
    <label>
      {label}
      {children}
    </label>
  ),
  Select: ({ options, value, onChange, placeholder, 'data-testid': testId }: any) => (
    <select
      data-testid={testId || 'select'}
      value={value || ''}
      onChange={(e: any) => {
        const opt = options.find((o: any) => o.value === e.target.value);
        if (opt) {
          onChange(opt);
        }
      }}
    >
      {placeholder && (
        <option value="" disabled>
          {placeholder}
        </option>
      )}
      {options.map((o: any) => (
        <option key={o.value} value={o.value}>
          {o.label}
        </option>
      ))}
    </select>
  ),
  MultiSelect: ({ options, value, onChange, placeholder }: any) => (
    <select
      data-testid="multi-select"
      multiple
      value={(value || []).map((v: any) => v.value)}
      onChange={(e: any) => {
        const selected = Array.from(e.target.selectedOptions).map((opt: any) => {
          return options.find((o: any) => o.value === opt.value);
        });
        onChange(selected);
      }}
    >
      {placeholder && <option disabled>{placeholder}</option>}
      {options.map((o: any) => (
        <option key={o.value} value={o.value}>
          {o.label}
        </option>
      ))}
    </select>
  ),
  Input: ({ value, onChange, placeholder, ...rest }: any) => (
    <input value={value || ''} onChange={onChange} placeholder={placeholder} {...rest} />
  ),
  Button: ({ children, onClick }: any) => <button onClick={onClick}>{children}</button>,
}));

function makeQuery(overrides: Partial<ArcadeDBQuery> = {}): ArcadeDBQuery {
  return {
    refId: 'A',
    queryMode: QueryMode.TimeSeries,
    rawQuery: '',
    ...overrides,
  };
}

function makeDatasource(metadataPromise?: Promise<ArcadeDBMetadata>) {
  return {
    getMetadata: jest.fn().mockReturnValue(metadataPromise || Promise.resolve(mockMetadata)),
  } as any;
}

describe('TimeSeriesEditor', () => {
  it('renders loading state while fetching metadata', () => {
    // Create a promise that never resolves during this test
    const slowPromise = new Promise<ArcadeDBMetadata>(() => {});
    const ds = makeDatasource(slowPromise);
    const query = makeQuery();

    render(
      <TimeSeriesEditor query={query} onChange={jest.fn()} onRunQuery={jest.fn()} datasource={ds} />
    );

    expect(screen.getByText('Loading metadata...')).toBeInTheDocument();
  });

  it('renders type selector after metadata loads', async () => {
    const ds = makeDatasource();
    const query = makeQuery();

    render(
      <TimeSeriesEditor query={query} onChange={jest.fn()} onRunQuery={jest.fn()} datasource={ds} />
    );

    await waitFor(() => {
      expect(screen.queryByText('Loading metadata...')).not.toBeInTheDocument();
    });

    // Should have a select with "Select type..." placeholder
    const selects = screen.getAllByTestId('select');
    const typeSelect = selects[0];
    expect(typeSelect).toBeInTheDocument();
    expect(typeSelect).toHaveTextContent('Select type...');
  });

  it('selecting a type updates query', async () => {
    const ds = makeDatasource();
    const query = makeQuery();
    const onChange = jest.fn();

    render(
      <TimeSeriesEditor query={query} onChange={onChange} onRunQuery={jest.fn()} datasource={ds} />
    );

    await waitFor(() => {
      expect(screen.queryByText('Loading metadata...')).not.toBeInTheDocument();
    });

    const selects = screen.getAllByTestId('select');
    const typeSelect = selects[0];
    fireEvent.change(typeSelect, { target: { value: 'weather' } });

    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({ tsType: 'weather', tsFields: [], tsTags: {} })
    );
  });

  it('renders tag inputs when type with tags is selected', async () => {
    const ds = makeDatasource();
    const query = makeQuery({ tsType: 'weather' });

    render(
      <TimeSeriesEditor query={query} onChange={jest.fn()} onRunQuery={jest.fn()} datasource={ds} />
    );

    await waitFor(() => {
      expect(screen.queryByText('Loading metadata...')).not.toBeInTheDocument();
    });

    expect(screen.getByText('Tag Filters')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('filter value')).toBeInTheDocument();
  });

  it('renders aggregation options', async () => {
    const ds = makeDatasource();
    const query = makeQuery();

    render(
      <TimeSeriesEditor query={query} onChange={jest.fn()} onRunQuery={jest.fn()} datasource={ds} />
    );

    await waitFor(() => {
      expect(screen.queryByText('Loading metadata...')).not.toBeInTheDocument();
    });

    expect(screen.getByText('Aggregation (optional)')).toBeInTheDocument();
    // Verify aggregation function select has options
    const selects = screen.getAllByTestId('select');
    // The aggregation function select should contain AVG, SUM, etc.
    const aggSelect = selects.find((s) => s.textContent?.includes('AVG'));
    expect(aggSelect).toBeDefined();
  });

  it('renders field selector after metadata loads', async () => {
    const ds = makeDatasource();
    const query = makeQuery();

    render(
      <TimeSeriesEditor query={query} onChange={jest.fn()} onRunQuery={jest.fn()} datasource={ds} />
    );

    await waitFor(() => {
      expect(screen.queryByText('Loading metadata...')).not.toBeInTheDocument();
    });

    expect(screen.getByTestId('multi-select')).toBeInTheDocument();
  });
});
