package pgquery

import (
	pg_query "github.com/pganalyze/pg_query_go/v5"
)

func GetCreateIndexStatement(statement *pg_query.RawStmt) *pg_query.IndexStmt {
	if statement == nil {
		return nil
	}
	return statement.Stmt.GetIndexStmt()
}

func GetDropIndexStatement(statement *pg_query.RawStmt) *pg_query.DropStmt {
	if statement == nil {
		return nil
	}
	drop := statement.Stmt.GetDropStmt()
	if drop == nil {
		return nil
	}
	if drop.RemoveType != pg_query.ObjectType_OBJECT_INDEX {
		return nil
	}
	return drop
}
