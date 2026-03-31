import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { ArcadeDBDataSourceOptions, ArcadeDBQuery } from './types';

export const plugin = new DataSourcePlugin<DataSource, ArcadeDBQuery, ArcadeDBDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
