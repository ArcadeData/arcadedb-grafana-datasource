import React from 'react';
import { CodeEditor } from '@grafana/ui';
import { ArcadeDBQuery } from '../types';

interface Props {
  query: ArcadeDBQuery;
  onChange: (query: ArcadeDBQuery) => void;
  onRunQuery: () => void;
}

export function CypherEditor({ query, onChange, onRunQuery }: Props) {
  const onQueryChange = (value: string) => {
    onChange({ ...query, rawQuery: value });
  };

  return (
    <div style={{ marginTop: 4 }}>
      <CodeEditor
        language="plaintext"
        value={query.rawQuery || ''}
        height={200}
        showMiniMap={false}
        showLineNumbers={true}
        onBlur={onQueryChange}
        onSave={(value) => {
          onQueryChange(value);
          onRunQuery();
        }}
      />
      <div style={{ marginTop: 4, fontSize: 12, color: '#8e8e8e' }}>
        OpenCypher query. Enable Node Graph toggle to visualize as a graph.
        &nbsp;|&nbsp; Macros: <code>$__timeFrom</code>, <code>$__timeTo</code>
        &nbsp;|&nbsp; Press Ctrl+S / Cmd+S to run
      </div>
    </div>
  );
}
