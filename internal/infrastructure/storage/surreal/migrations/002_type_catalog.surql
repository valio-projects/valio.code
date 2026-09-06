BEGIN TRANSACTION;
DEFINE TABLE IF NOT EXISTS project_revision SCHEMALESS;
DEFINE TABLE IF NOT EXISTS type_descriptor SCHEMALESS;
DEFINE TABLE IF NOT EXISTS manifest_shard SCHEMALESS;
DEFINE TABLE IF NOT EXISTS search_block SCHEMALESS;
DEFINE INDEX IF NOT EXISTS project_definition_key ON project FIELDS payload.project.key UNIQUE;
DEFINE INDEX IF NOT EXISTS type_name ON type_descriptor FIELDS payload.name;
UPSERT schema_migration:type_catalog SET version = 2;
COMMIT TRANSACTION;
