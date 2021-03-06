package record

import (
	"errors"
	"strings"

	"github.com/IacopoMelani/Go-Starter-Project/pkg/manager/db"
	builder "github.com/IacopoMelani/Go-Starter-Project/pkg/manager/db/query_builder"
	"github.com/jmoiron/sqlx"
)

// NewTableModel - Tipo per definire una funzione che restituisce una TableRecordInterface
type NewTableModel func() TableRecordInterface

// TableRecordInterface - interfaccia che definisce una generica struct che permette l'interazione con TableRecord
type TableRecordInterface interface {
	GetTableRecord() *TableRecord
	GetPrimaryKeyName() string
	GetPrimaryKeyValue() int64
	GetTableName() string
}

// TableRecord - Struct per l'implementazione di TableRecordInterface
// implementa QueryBuilderInterface
type TableRecord struct {
	isNew      bool
	isReadOnly bool
	builder.Builder
	db db.SQLConnector
}

// getTableRecordConnection - Restituisce la connessione di un TableRecordInterface
func getTableRecordConnection(ti TableRecordInterface) db.SQLConnector {
	return ti.GetTableRecord().db
}

// save - Si occupa di inserire un nuovo record nella tabella
func save(ti TableRecordInterface) error {

	t := ti.GetTableRecord()

	query := genSaveQuery(ti)
	fValue := getFieldsValueNoPrimary(ti)
	id, err := t.executeSaveUpdateQuery(query, fValue)
	if err != nil {
		return err
	}

	if err := LoadByID(ti, id); err != nil {
		return err
	}

	t.SetIsNew(false)

	return nil
}

// update - Si occupa di aggiornare il record nel database
func update(ti TableRecordInterface) error {

	t := ti.GetTableRecord()

	query := genUpdateQuery(ti)
	fValue := getFieldsValueNoPrimary(ti)
	_, err := t.executeSaveUpdateQuery(query, append(fValue, ti.GetPrimaryKeyValue()))
	if err != nil {
		return err
	}

	if err := LoadByID(ti, ti.GetPrimaryKeyValue()); err != nil {
		return err
	}

	return nil
}

// executeSaveUpdateQuery - Si occupa di eseguire fisicamente la query, in caso di successo restituisce l'Id appena inserito
func (t *TableRecord) executeSaveUpdateQuery(query string, params []interface{}) (int64, error) {

	db := t.db

	res, err := db.Exec(query, params...)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastID, nil
}

// AllField - Restitusice tutti i campi per la select *
func AllField(ti TableRecordInterface) string {

	fieldName, _ := GetFieldMapper(ti)

	return strings.Join(fieldName, ",")
}

// All - Restituisce tutti i risultati per il costruttore del table record passato
func All(ntm NewTableModel) ([]TableRecordInterface, error) {

	var result []TableRecordInterface

	pivot := ntm()

	db := pivot.GetTableRecord().db

	query := "SELECT " + AllField(pivot) + " FROM " + pivot.GetTableName()

	rows, err := db.Queryx(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {

		ti := ntm()

		err = LoadFromRow(rows, ti)
		if err != nil {
			return nil, err
		}

		result = append(result, ti)

	}

	return result, nil
}

// Delete - Si occupa di cancellare un record sul database
func Delete(ti TableRecordInterface) (int64, error) {

	db := getTableRecordConnection(ti)

	stmt, err := db.Prepare(genDeleteQuery(ti))
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(ti.GetPrimaryKeyValue())
	if err != nil {
		return 0, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, nil
	}

	return rows, nil
}

// ExecQuery - Esegue la query costruita con QueryBuilder
func ExecQuery(ti TableRecordInterface, ntm NewTableModel) ([]TableRecordInterface, error) {

	t := ti.GetTableRecord()

	stmt, err := t.PrepareStmt(ti.GetTableName())
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Queryx(t.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiList []TableRecordInterface

	for rows.Next() {

		nti := ntm()

		if err := LoadFromRow(rows, nti); err != nil {
			return nil, err
		}

		tiList = append(tiList, nti)
	}

	ti.GetTableRecord().ResetStmt()

	return tiList, nil
}

// LoadByID - Carica l'istanza passata con i valori della sua tabella ricercando per chiave primaria
func LoadByID(ti TableRecordInterface, id int64) error {

	db := getTableRecordConnection(ti)

	query := "SELECT " + AllField(ti) + " FROM " + ti.GetTableName() + " WHERE " + ti.GetPrimaryKeyName() + " = ?"

	params := []interface{}{interface{}(id)}

	stmt, err := db.Preparex(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Queryx(params...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {

		if err := LoadFromRow(rows, ti); err != nil {
			return err
		}
	}

	return nil
}

// LoadFromRow - Si occupa di caricare la struct dal result - row della query
func LoadFromRow(r *sqlx.Rows, tri TableRecordInterface) error {

	if err := r.StructScan(tri); err != nil {
		return err
	}

	tri.GetTableRecord().SetIsNew(false).SetSQLConnection(tri.GetTableRecord().db)

	return nil
}

// NewTableRecord - Restituisce una nuova istanza di TableRecord
func NewTableRecord(isNew bool, isReadOnly bool) *TableRecord {

	tr := new(TableRecord)
	tr.isNew = isNew
	tr.isReadOnly = isReadOnly

	return tr
}

// Save - Si occupa di eseguire il salvataggio della TableRecord eseguendo un inserimento se TableRecord::isNew risulta false, altrimenti ne aggiorna il valore
func Save(ti TableRecordInterface) error {

	t := ti.GetTableRecord()

	if t.isReadOnly {
		return errors.New("Read-only model")
	}

	if t.isNew {

		err := save(ti)
		if err != nil {
			return err
		}

	} else {

		err := update(ti)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetDB - Restituisce la risorsa di connessione al database
func (t *TableRecord) GetDB() db.SQLConnector {
	return t.db
}

// IsNew - Restituisce se il record è nuovo
func (t *TableRecord) IsNew() bool {
	return t.isNew
}

// PrepareStmt - Restituisce lo stmt della query pronta da essere eseguita
func (t *TableRecord) PrepareStmt(tableName string) (*sqlx.Stmt, error) {

	db := t.db

	query := t.BuildQuery(tableName)

	stmt, err := db.Preparex(query)
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

// SetIsNew - Si occupa di impostare il valore del campo TableRecord::isNews
func (t *TableRecord) SetIsNew(new bool) *TableRecord {
	t.isNew = new
	return t
}

// SetSQLConnection - Imposta la connessione
func (t *TableRecord) SetSQLConnection(db db.SQLConnector) *TableRecord {
	t.db = db
	return t
}
