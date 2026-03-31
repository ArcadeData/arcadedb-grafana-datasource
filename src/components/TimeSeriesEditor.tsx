import React, { useEffect, useState } from 'react';
import { InlineField, Select, MultiSelect, Input, Button } from '@grafana/ui';
import { SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { ArcadeDBQuery, ArcadeDBMetadata } from '../types';

interface Props {
  query: ArcadeDBQuery;
  onChange: (query: ArcadeDBQuery) => void;
  onRunQuery: () => void;
  datasource: DataSource;
}

const aggregationOptions: Array<SelectableValue<string>> = [
  { label: 'AVG', value: 'AVG' },
  { label: 'SUM', value: 'SUM' },
  { label: 'MIN', value: 'MIN' },
  { label: 'MAX', value: 'MAX' },
  { label: 'COUNT', value: 'COUNT' },
];

export function TimeSeriesEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const [metadata, setMetadata] = useState<ArcadeDBMetadata | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    datasource
      .getMetadata()
      .then(setMetadata)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [datasource]);

  const selectedType = metadata?.types.find((t) => t.name === query.tsType);

  const typeOptions: Array<SelectableValue<string>> =
    metadata?.types.map((t) => ({ label: t.name, value: t.name })) || [];

  const fieldOptions: Array<SelectableValue<string>> =
    selectedType?.fields.map((f) => ({ label: `${f.name} (${f.dataType})`, value: f.name })) || [];

  const tagOptions =
    selectedType?.tags.map((t) => ({ label: t.name, value: t.name })) || [];

  const onTypeChange = (value: string) => {
    onChange({ ...query, tsType: value, tsFields: [], tsTags: {} });
  };

  const onFieldsChange = (values: Array<SelectableValue<string>>) => {
    onChange({ ...query, tsFields: values.map((v) => v.value!) });
  };

  const onTagChange = (tagName: string, tagValue: string) => {
    onChange({ ...query, tsTags: { ...query.tsTags, [tagName]: tagValue } });
  };

  const onAggregationChange = (field: string, type: string, alias: string, bucketInterval: number) => {
    onChange({
      ...query,
      tsAggregation: {
        bucketInterval: bucketInterval || undefined,
        requests: [{ field, type: type as any, alias: alias || undefined }],
      },
    });
  };

  if (loading) {
    return <div>Loading metadata...</div>;
  }

  const aggRequest = query.tsAggregation?.requests?.[0];

  return (
    <>
      <InlineField label="Type" labelWidth={16} tooltip="Time series type">
        <Select
          options={typeOptions}
          value={query.tsType}
          onChange={(v) => onTypeChange(v.value!)}
          width={30}
          placeholder="Select type..."
        />
      </InlineField>

      <InlineField label="Fields" labelWidth={16} tooltip="Fields to query">
        <MultiSelect
          options={fieldOptions}
          value={query.tsFields?.map((f) => ({ label: f, value: f }))}
          onChange={onFieldsChange}
          width={40}
          placeholder="All fields"
        />
      </InlineField>

      {tagOptions.length > 0 && (
        <>
          <h6 style={{ marginTop: 8 }}>Tag Filters</h6>
          {tagOptions.map((tag) => (
            <InlineField key={tag.value} label={tag.label!} labelWidth={16}>
              <Input
                width={30}
                value={query.tsTags?.[tag.value!] || ''}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => onTagChange(tag.value!, e.target.value)}
                placeholder="filter value"
              />
            </InlineField>
          ))}
        </>
      )}

      <h6 style={{ marginTop: 8 }}>Aggregation (optional)</h6>
      <div style={{ display: 'flex', gap: 8, alignItems: 'flex-end' }}>
        <InlineField label="Function" labelWidth={16}>
          <Select
            options={aggregationOptions}
            value={aggRequest?.type}
            onChange={(v) =>
              onAggregationChange(
                aggRequest?.field || query.tsFields?.[0] || '',
                v.value!,
                aggRequest?.alias || '',
                query.tsAggregation?.bucketInterval || 0
              )
            }
            width={16}
            placeholder="None"
            isClearable
          />
        </InlineField>

        <InlineField label="Field" labelWidth={10}>
          <Select
            options={fieldOptions}
            value={aggRequest?.field}
            onChange={(v) =>
              onAggregationChange(
                v.value!,
                aggRequest?.type || 'AVG',
                aggRequest?.alias || '',
                query.tsAggregation?.bucketInterval || 0
              )
            }
            width={20}
          />
        </InlineField>

        <InlineField label="Bucket (ms)" labelWidth={16}>
          <Input
            type="number"
            width={16}
            value={query.tsAggregation?.bucketInterval || ''}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
              onAggregationChange(
                aggRequest?.field || '',
                aggRequest?.type || 'AVG',
                aggRequest?.alias || '',
                parseInt(e.target.value, 10) || 0
              )
            }
            placeholder="auto"
          />
        </InlineField>

        <Button variant="primary" size="sm" onClick={onRunQuery}>
          Run
        </Button>
      </div>
    </>
  );
}
