import React from 'react';
import { render, screen } from '@testing-library/react';
import { SqlEditor } from './SqlEditor';
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
    onRunQuery: jest.fn(),
  };
}

describe('SqlEditor', () => {
  beforeEach(() => { capturedOnBlur = undefined; });

  test('renders code editor with current query', () => {
    render(<SqlEditor {...makeProps('SELECT * FROM Person')} />);
    expect(screen.getByTestId('code-editor')).toHaveValue('SELECT * FROM Person');
  });

  test('calls onChange on blur', () => {
    const props = makeProps('SELECT 1');
    render(<SqlEditor {...props} />);
    capturedOnBlur!('SELECT 2');
    expect(props.onChange).toHaveBeenCalledWith(expect.objectContaining({ rawQuery: 'SELECT 2' }));
  });
});
