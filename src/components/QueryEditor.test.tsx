import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryEditor } from './QueryEditor';
import { QueryMode, ArcadeDBQuery } from '../types';

jest.mock('./TimeSeriesEditor', () => ({ TimeSeriesEditor: () => <div data-testid="timeseries-editor" /> }));
jest.mock('./SqlEditor', () => ({ SqlEditor: () => <div data-testid="sql-editor" /> }));
jest.mock('./CypherEditor', () => ({ CypherEditor: () => <div data-testid="cypher-editor" /> }));
jest.mock('./GremlinEditor', () => ({ GremlinEditor: () => <div data-testid="gremlin-editor" /> }));

jest.mock('@grafana/ui', () => ({
  InlineField: ({ label, children }: any) => <label>{label}{children}</label>,
  Select: ({ options, value, onChange }: any) => (
    <select
      data-testid="mode-select"
      value={value}
      onChange={(e: any) => {
        const opt = options.find((o: any) => o.value === e.target.value);
        if (opt) {
          onChange(opt);
        }
      }}
    >
      {options.map((o: any) => (
        <option key={o.value} value={o.value}>
          {o.label}
        </option>
      ))}
    </select>
  ),
  InlineSwitch: ({ value, onChange }: any) => (
    <input type="checkbox" data-testid="node-graph-toggle" checked={value} onChange={onChange} />
  ),
}));

function makeProps(overrides: Partial<ArcadeDBQuery> = {}) {
  const query: ArcadeDBQuery = {
    refId: 'A',
    queryMode: QueryMode.SQL,
    rawQuery: '',
    ...overrides,
  };
  return {
    query,
    onChange: jest.fn(),
    onRunQuery: jest.fn(),
    datasource: {} as any,
  };
}

describe('QueryEditor', () => {
  it('renders with default SQL mode', () => {
    const props = makeProps();
    render(<QueryEditor {...props} />);

    expect(screen.getByTestId('sql-editor')).toBeInTheDocument();
    expect(screen.queryByTestId('cypher-editor')).not.toBeInTheDocument();
    expect(screen.queryByTestId('gremlin-editor')).not.toBeInTheDocument();
    expect(screen.queryByTestId('timeseries-editor')).not.toBeInTheDocument();
  });

  it('switches to OpenCypher mode', () => {
    const props = makeProps();
    render(<QueryEditor {...props} />);

    fireEvent.change(screen.getByTestId('mode-select'), { target: { value: QueryMode.Cypher } });

    expect(props.onChange).toHaveBeenCalledWith(
      expect.objectContaining({ queryMode: QueryMode.Cypher })
    );
  });

  it('switches to Gremlin mode', () => {
    const props = makeProps();
    render(<QueryEditor {...props} />);

    fireEvent.change(screen.getByTestId('mode-select'), { target: { value: QueryMode.Gremlin } });

    expect(props.onChange).toHaveBeenCalledWith(
      expect.objectContaining({ queryMode: QueryMode.Gremlin })
    );
  });

  it('switches to TimeSeries mode', () => {
    const props = makeProps();
    render(<QueryEditor {...props} />);

    fireEvent.change(screen.getByTestId('mode-select'), { target: { value: QueryMode.TimeSeries } });

    expect(props.onChange).toHaveBeenCalledWith(
      expect.objectContaining({ queryMode: QueryMode.TimeSeries })
    );
  });

  it('shows node graph toggle for Cypher mode', () => {
    const props = makeProps({ queryMode: QueryMode.Cypher });
    render(<QueryEditor {...props} />);

    expect(screen.getByTestId('node-graph-toggle')).toBeInTheDocument();
  });

  it('hides node graph toggle for timeseries mode', () => {
    const props = makeProps({ queryMode: QueryMode.TimeSeries });
    render(<QueryEditor {...props} />);

    expect(screen.queryByTestId('node-graph-toggle')).not.toBeInTheDocument();
  });
});
