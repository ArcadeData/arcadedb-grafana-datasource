import React from 'react';
import { render, screen } from '@testing-library/react';
import { VariableQueryEditor } from './VariableQueryEditor';
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
    query: { refId: 'A', queryMode: QueryMode.SQL, rawQuery } as ArcadeDBQuery,
    onChange: jest.fn(),
  };
}

describe('VariableQueryEditor', () => {
  beforeEach(() => { capturedOnBlur = undefined; });

  test('renders with query text', () => {
    render(<VariableQueryEditor {...makeProps('SELECT name FROM types')} />);
    expect(screen.getByTestId('code-editor')).toHaveValue('SELECT name FROM types');
  });

  test('calls onChange when query changes', () => {
    const props = makeProps('SELECT 1');
    render(<VariableQueryEditor {...props} />);
    capturedOnBlur!('SELECT name FROM Person');
    expect(props.onChange).toHaveBeenCalledWith(
      expect.objectContaining({ rawQuery: 'SELECT name FROM Person' }),
      'SELECT name FROM Person'
    );
  });
});
