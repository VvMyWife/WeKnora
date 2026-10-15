-- Only run after exporting originals; normal application rollback keeps tables.
DROP TABLE IF EXISTS archive_mapping_links;
DROP TABLE IF EXISTS archive_mapping_items;
DROP TABLE IF EXISTS archive_mapping_versions;
DROP TABLE IF EXISTS archive_mapping_batches;
