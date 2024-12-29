package sqld

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// AutoRegisterEnums queries the current schema for all user-defined enums,
// then registers them in pgx's TypeMap as text so we avoid unknown OID errors.
//
// Note on Schema Behavior:
// - current_schema() returns the first schema in the search_path that exists
// - By default, this is typically 'public' unless explicitly changed
// - The search_path can be modified at the database, user, or session level
//
// Future Enhancements:
// - Add schema parameter to support multi-schema databases
// - Support array of schemas to register enums from multiple schemas
// - Add option to specify search_path explicitly
//
// Example schema configurations:
//   SET search_path TO custom_schema, public;
//   ALTER DATABASE dbname SET search_path TO custom_schema, public;
//   ALTER USER username SET search_path TO custom_schema, public;
func AutoRegisterEnums(ctx context.Context, conn *pgx.Conn) error {
	rows, err := conn.Query(ctx, `
		SELECT t.oid, t.typname
		FROM pg_type t
		JOIN pg_namespace n ON t.typnamespace = n.oid
		WHERE t.typtype = 'e'
		  AND n.nspname = current_schema()
	`)
	if err != nil {
		return fmt.Errorf("failed to query pg_type for enums: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var oid uint32
		var typname string
		if scanErr := rows.Scan(&oid, &typname); scanErr != nil {
			return fmt.Errorf("failed to scan row for enum: %w", scanErr)
		}

		// Register the enum as text
		conn.TypeMap().RegisterType(&pgtype.Type{
			Name:  typname,
			OID:   oid,
			Codec: pgtype.TextCodec{},
		})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows iteration error: %w", err)
	}

	return nil
}
