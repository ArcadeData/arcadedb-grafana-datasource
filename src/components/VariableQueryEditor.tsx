import React from 'react';
import { CodeEditor } from '@grafana/ui';
import { ArcadeDBQuery } from '../types';

interface Props {
  query: ArcadeDBQuery;
  onChange: (query: ArcadeDBQuery, definition: string) => void;
}

export function VariableQueryEditor({ query, onChange }: Props) {
  const onQueryChange = (value: string) => {
    onChange({ ...query, rawQuery: value }, value);
  };

  return (
    <div>
      <CodeEditor
        language="sql"
        value={query.rawQuery || ''}
        height={100}
        showMiniMap={false}
        showLineNumbers={true}
        onBlur={onQueryChange}
      />
      <div style={{ marginTop: 4, fontSize: 12, color: '#8e8e8e' }}>
        Enter a SQL query that returns a single column. Values will populate the template variable.
      </div>
    </div>
  );
}
