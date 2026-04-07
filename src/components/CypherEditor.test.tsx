import React from 'react';
import { render, screen } from '@testing-library/react';
import { CypherEditor } from './CypherEditor';
import { QueryMode, ArcadeDBQuery } from '../types';

let capturedOnBlur: ((value: string) => void) | undefined;

jest.mock('@grafana/ui', () => ({
  CodeEditor: ({ value, onBlur }: any) => {
    capturedOnBlur = onBlur;
    return <textarea data-testid="code-editor" defaultValue={value} />;
  },
}));

function makeProps(rawQuery = '') {
  return {
    query: { refId: 'A', queryMode: QueryMode.Cypher, rawQuery } as ArcadeDBQuery,
    onChange: jest.fn(),
    onRunQuery: jest.fn(),
  };
}

describe('CypherEditor', () => {
  beforeEach(() => { capturedOnBlur = undefined; });

  test('renders code editor with current query', () => {
    render(<CypherEditor {...makeProps('MATCH (n) RETURN n')} />);
    expect(screen.getByTestId('code-editor')).toHaveValue('MATCH (n) RETURN n');
  });

  test('calls onChange on blur', () => {
    const props = makeProps('MATCH (n) RETURN n');
    render(<CypherEditor {...props} />);
    capturedOnBlur!('MATCH (p:Person) RETURN p');
    expect(props.onChange).toHaveBeenCalledWith(expect.objectContaining({ rawQuery: 'MATCH (p:Person) RETURN p' }));
  });
});
