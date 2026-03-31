import React from 'react';
import { QueryEditorProps } from '@grafana/data';
import { InlineField, Select, InlineSwitch } from '@grafana/ui';
import { DataSource } from '../datasource';
import { ArcadeDBDataSourceOptions, ArcadeDBQuery, QueryMode } from '../types';
import { TimeSeriesEditor } from './TimeSeriesEditor';
import { SqlEditor } from './SqlEditor';
import { CypherEditor } from './CypherEditor';
import { GremlinEditor } from './GremlinEditor';

type Props = QueryEditorProps<DataSource, ArcadeDBQuery, ArcadeDBDataSourceOptions>;

const queryModeOptions = [
  { label: 'Time Series', value: QueryMode.TimeSeries },
  { label: 'SQL', value: QueryMode.SQL },
  { label: 'Cypher', value: QueryMode.Cypher },
  { label: 'Gremlin', value: QueryMode.Gremlin },
];

export function QueryEditor(props: Props) {
  const { query, onChange, onRunQuery, datasource } = props;
  const queryMode = query.queryMode || QueryMode.SQL;
  const showNodeGraphToggle = queryMode === QueryMode.Cypher || queryMode === QueryMode.Gremlin;

  const onQueryModeChange = (value: QueryMode) => {
    onChange({ ...query, queryMode: value });
  };

  const onNodeGraphToggle = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, nodeGraphEnabled: event.currentTarget.checked });
    onRunQuery();
  };

  return (
    <>
      <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
        <InlineField label="Mode" labelWidth={10}>
          <Select
            options={queryModeOptions}
            value={queryMode}
            onChange={(v) => onQueryModeChange(v.value!)}
            width={20}
          />
        </InlineField>

        {showNodeGraphToggle && (
          <InlineField label="Node Graph" tooltip="Render results as a graph using the Node Graph panel">
            <InlineSwitch
              value={query.nodeGraphEnabled || false}
              onChange={onNodeGraphToggle}
            />
          </InlineField>
        )}
      </div>

      {queryMode === QueryMode.TimeSeries && (
        <TimeSeriesEditor query={query} onChange={onChange} onRunQuery={onRunQuery} datasource={datasource} />
      )}
      {queryMode === QueryMode.SQL && (
        <SqlEditor query={query} onChange={onChange} onRunQuery={onRunQuery} />
      )}
      {queryMode === QueryMode.Cypher && (
        <CypherEditor query={query} onChange={onChange} onRunQuery={onRunQuery} />
      )}
      {queryMode === QueryMode.Gremlin && (
        <GremlinEditor query={query} onChange={onChange} onRunQuery={onRunQuery} />
      )}
    </>
  );
}
