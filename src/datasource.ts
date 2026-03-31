import {
  DataSourceInstanceSettings,
  CoreApp,
  ScopedVars,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { ArcadeDBDataSourceOptions, ArcadeDBQuery, ArcadeDBMetadata, defaultQuery } from './types';

export class DataSource extends DataSourceWithBackend<ArcadeDBQuery, ArcadeDBDataSourceOptions> {
  database: string;

  constructor(instanceSettings: DataSourceInstanceSettings<ArcadeDBDataSourceOptions>) {
    super(instanceSettings);
    this.database = instanceSettings.jsonData.database || '';
  }

  getDefaultQuery(_: CoreApp): Partial<ArcadeDBQuery> {
    return defaultQuery;
  }

  applyTemplateVariables(query: ArcadeDBQuery, scopedVars: ScopedVars): ArcadeDBQuery {
    const templateSrv = getTemplateSrv();
    return {
      ...query,
      rawQuery: templateSrv.replace(query.rawQuery, scopedVars),
    };
  }

  async getMetadata(): Promise<ArcadeDBMetadata> {
    return this.getResource('metadata');
  }

  filterQuery(query: ArcadeDBQuery): boolean {
    return !query.hide;
  }
}
