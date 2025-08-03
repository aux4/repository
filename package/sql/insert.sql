INSERT INTO TABLE_NAME (id, data, metadata) VALUES ('KEY', 'JSON_DATA', 'METADATA')
ON CONFLICT(id) DO UPDATE SET 
  data = excluded.data,
  metadata = excluded.metadata,
  updated_at = CURRENT_TIMESTAMP
