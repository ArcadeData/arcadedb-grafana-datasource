import { DataQuery, DataSourceJsonData } from '@grafana/data';

export enum QueryMode {
  TimeSeries = 'timeseries',
  SQL = 'sql',
  Cypher = 'cypher',
  Gremlin = 'gremlin',
}

export interface ArcadeDBQuery extends DataQuery {
  queryMode: QueryMode;
  rawQuery: string;

  // Time Series mode fields
  tsType?: string;
  tsFields?: string[];
  tsTags?: Record<string, string>;
  tsAggregation?: {
    bucketInterval?: number;
    requests?: Array<{
      field: string;
      type: 'SUM' | 'AVG' | 'MIN' | 'MAX' | 'COUNT';
      alias?: string;
    }>;
  };

  // Graph mode
  nodeGraphEnabled?: boolean;
}

export const defaultQuery: Partial<ArcadeDBQuery> = {
  queryMode: QueryMode.SQL,
  rawQuery: '',
  nodeGraphEnabled: false,
};

export interface ArcadeDBDataSourceOptions extends DataSourceJsonData {
  database: string;
}

export interface ArcadeDBSecureJsonData {
  password?: string;
  basicAuthPassword?: string;
}

// Metadata response from /grafana/metadata
export interface ArcadeDBMetadata {
  types: ArcadeDBTimeSeriesType[];
  aggregationTypes: string[];
}

export interface ArcadeDBTimeSeriesType {
  name: string;
  fields: Array<{ name: string; dataType: string }>;
  tags: Array<{ name: string; dataType: string }>;
}
