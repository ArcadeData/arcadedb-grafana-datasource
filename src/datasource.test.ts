import { DataSource } from './datasource';
import { QueryMode, ArcadeDBQuery } from './types';
import { CoreApp, DataSourceInstanceSettings } from '@grafana/data';
import { getTemplateSrv } from '@grafana/runtime';

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  DataSourceWithBackend: class {
    constructor() {}
    getResource = jest.fn();
  },
  getTemplateSrv: jest.fn(),
}));

function createDataSource(): DataSource {
  const settings = {
    id: 1,
    uid: 'test',
    type: 'arcadedb-arcadedb-datasource',
    name: 'ArcadeDB',
    url: 'http://localhost:2480',
    jsonData: { database: 'testdb' },
    access: 'proxy',
    meta: {} as any,
    readOnly: false,
  } as DataSourceInstanceSettings<any>;

  return new DataSource(settings);
}

describe('DataSource', () => {
  test('getDefaultQuery returns SQL mode with empty query', () => {
    const ds = createDataSource();
    const defaultQuery = ds.getDefaultQuery(CoreApp.PanelEditor);
    expect(defaultQuery.queryMode).toBe(QueryMode.SQL);
    expect(defaultQuery.rawQuery).toBe('');
    expect(defaultQuery.nodeGraphEnabled).toBe(false);
  });

  test('filterQuery includes non-hidden queries', () => {
    const ds = createDataSource();
    const query = { refId: 'A', queryMode: QueryMode.SQL, rawQuery: 'SELECT 1', hide: false } as ArcadeDBQuery;
    expect(ds.filterQuery(query)).toBe(true);
  });

  test('filterQuery excludes hidden queries', () => {
    const ds = createDataSource();
    const query = { refId: 'A', queryMode: QueryMode.SQL, rawQuery: 'SELECT 1', hide: true } as ArcadeDBQuery;
    expect(ds.filterQuery(query)).toBe(false);
  });

  test('applyTemplateVariables calls templateSrv.replace', () => {
    const replaceMock = jest.fn().mockReturnValue('SELECT * FROM resolved');
    (getTemplateSrv as jest.Mock).mockReturnValue({ replace: replaceMock });

    const ds = createDataSource();
    const query = {
      refId: 'A',
      queryMode: QueryMode.SQL,
      rawQuery: 'SELECT * FROM $variable',
    } as ArcadeDBQuery;

    const result = ds.applyTemplateVariables(query, {});
    expect(replaceMock).toHaveBeenCalledWith('SELECT * FROM $variable', {});
    expect(result.rawQuery).toBe('SELECT * FROM resolved');
  });
});
