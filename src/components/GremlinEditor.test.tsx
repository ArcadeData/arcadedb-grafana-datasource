import React from 'react';
import { render, screen } from '@testing-library/react';
import { GremlinEditor } from './GremlinEditor';
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
    query: { refId: 'A', queryMode: QueryMode.Gremlin, rawQuery } as ArcadeDBQuery,
    onChange: jest.fn(),
    onRunQuery: jest.fn(),
  };
}

describe('GremlinEditor', () => {
  beforeEach(() => { capturedOnBlur = undefined; });

  test('renders code editor with current query', () => {
    render(<GremlinEditor {...makeProps('g.V().count()')} />);
    expect(screen.getByTestId('code-editor')).toHaveValue('g.V().count()');
  });

  test('calls onChange on blur', () => {
    const props = makeProps('g.V().count()');
    render(<GremlinEditor {...props} />);
    capturedOnBlur!('g.V().hasLabel("Person")');
    expect(props.onChange).toHaveBeenCalledWith(expect.objectContaining({ rawQuery: 'g.V().hasLabel("Person")' }));
  });
});
