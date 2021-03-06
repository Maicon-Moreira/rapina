package parsers

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

const currentDbVersion = 200301

var createTableMap = map[string]string{
	"dfp": `CREATE TABLE IF NOT EXISTS dfp
	(
		"ID" PRIMARY KEY,
		"ID_CIA" integer,
		"CODE" integer,
		"YEAR" string,
		"DATA_TYPE" string,

		"VERSAO" integer,
		"MOEDA" varchar(4),
		"ESCALA_MOEDA" varchar(7),
		"DT_FIM_EXERC" integer,
		"CD_CONTA" varchar(18),
		"DS_CONTA" varchar(100),
		"VL_CONTA" real
	);`,

	"itr": `CREATE TABLE IF NOT EXISTS itr
	(
		"ID" PRIMARY KEY,
		"ID_CIA" integer,
		"CODE" integer,
		"YEAR" string,
		"DATA_TYPE" string,

		"VERSAO" integer,
		"MOEDA" varchar(4),
		"ESCALA_MOEDA" varchar(7),
		"DT_FIM_EXERC" integer,
		"CD_CONTA" varchar(18),
		"DS_CONTA" varchar(100),
		"VL_CONTA" real
	);`,

	"codes": `CREATE TABLE IF NOT EXISTS codes
	(
		"CODE" INTEGER NOT NULL PRIMARY KEY,
		"NAME" varchar(100)
	);`,

	"companies": `CREATE TABLE IF NOT EXISTS companies
	(
		"ID" INTEGER NOT NULL PRIMARY KEY,
		"CNPJ" varchar(20),
		"NAME" varchar(100)
	);`,

	"md5": `CREATE TABLE IF NOT EXISTS md5
	(
		md5 NOT NULL PRIMARY KEY
	);`,

	"status": `CREATE TABLE IF NOT EXISTS status
	(
		table_name TEXT NOT NULL PRIMARY KEY,
		version integer
	);`,
}

//
// whatTable for the data type
//
func whatTable(dataType string) (table string, err error) {
	switch dataType {
	case "BPA", "BPP", "DRE", "DFC_MD", "DFC_MI", "DVA":
		table = "dfp"
	case "BPA_ITR", "BPP_ITR", "DRE_ITR", "DFC_MD_ITR", "DFC_MI_ITR", "DVA_ITR":
		table = "itr"
	case "CODES":
		table = "codes"
	case "MD5":
		table = "md5"
	case "STATUS":
		table = "status"
	case "COMPANIES":
		table = "companies"
	default:
		return "", errors.Wrapf(err, "tipo de informação inexistente: %s", dataType)
	}

	return
}

//
// createTable creates the table if not created yet
//
func createTable(db *sql.DB, dataType string) (err error) {

	table, err := whatTable(dataType)
	if err != nil {
		return err
	}

	_, err = db.Exec(createTableMap[table])
	if err != nil {
		return errors.Wrap(err, "erro ao criar tabela "+table)
	}

	err = createIndexes(db, table)
	if err != nil {
		return errors.Wrap(err, "erro ao criar índice para table "+table)
	}

	if dataType != "STATUS" {
		_, err = db.Exec(fmt.Sprintf(`INSERT OR REPLACE INTO status (table_name, version) VALUES ("%s",%d)`, table, currentDbVersion))
		if err != nil {
			return errors.Wrap(err, "erro ao atualizar tabela "+table)
		}
	}

	return nil
}

//
// dbVersion returns the version stored in DB
//
func dbVersion(db *sql.DB, dataType string) (v int, table string) {
	table, err := whatTable(dataType)
	if err != nil {
		return
	}

	sqlStmt := `SELECT version FROM status WHERE table_name = ?`
	err = db.QueryRow(sqlStmt, table).Scan(&v)
	if err != nil {
		return
	}

	return
}

//
// wipeDB drops the table! Use with care
//
func wipeDB(db *sql.DB, dataType string) (err error) {
	table, err := whatTable(dataType)
	if err != nil {
		return
	}

	_, err = db.Exec("DROP TABLE IF EXISTS " + table)
	if err != nil {
		return errors.Wrap(err, "erro ao apagar tabela")
	}

	return
}

//
// createIndexes create indexes based on table name
//
func createIndexes(db *sql.DB, table string) error {
	indexes := []string{}

	switch table {
	case "dfp":
		indexes = []string{
			"CREATE INDEX IF NOT EXISTS dfp_metrics ON dfp (CODE, ID_CIA, YEAR, VL_CONTA);",
			"CREATE INDEX IF NOT EXISTS dfp_year_ver ON dfp (ID_CIA, YEAR, VERSAO);",
		}
	case "itr":
		indexes = []string{
			"CREATE INDEX IF NOT EXISTS itr_metrics ON itr (CODE, ID_CIA, YEAR, VL_CONTA);",
			"CREATE INDEX IF NOT EXISTS itr_quarter_ver ON itr (ID_CIA, DT_FIM_EXERC, VERSAO);",
		}
	}

	for _, idx := range indexes {
		_, err := db.Exec(idx)
		if err != nil {
			return errors.Wrap(err, "erro ao criar índice")
		}
	}

	return nil
}
