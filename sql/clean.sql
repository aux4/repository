SELECT 'DELETE FROM "' || name || '" WHERE json_extract(metadata, ''$.ttl'') IS NOT NULL AND json_extract(metadata, ''$.ttl'') <= strftime(''%s'', ''now'') + 0; SELECT ''' || name || ': '' || changes() || '' expired records removed'';'
FROM sqlite_master
WHERE type = 'table' AND name NOT LIKE 'sqlite_%';
