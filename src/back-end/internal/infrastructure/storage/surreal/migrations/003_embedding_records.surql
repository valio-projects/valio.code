BEGIN TRANSACTION;
DEFINE TABLE IF NOT EXISTS embedding_record SCHEMAFULL;
DEFINE FIELD IF NOT EXISTS key ON embedding_record TYPE string;
DEFINE FIELD IF NOT EXISTS payload ON embedding_record TYPE object FLEXIBLE;
DEFINE INDEX IF NOT EXISTS embedding_scope ON embedding_record FIELDS payload.workspaceId, payload.viewId, payload.profileFingerprint;
UPSERT schema_migration:embedding_records SET version = 3;
COMMIT TRANSACTION;
