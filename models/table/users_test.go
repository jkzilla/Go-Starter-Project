package table

import (
	"testing"

	"github.com/IacopoMelani/Go-Starter-Project/pkg/helpers/copy"

	record "github.com/IacopoMelani/Go-Starter-Project/pkg/models/table_record"
	"github.com/subosito/gotenv"
)

func TestTableMirror(t *testing.T) {

	gotenv.Load("./../../.env")

	u := NewUser()

	u.Name = copy.String("Mario")
	u.Lastname = copy.String("Rossi")
	u.Gender = copy.String("M")

	err := record.Save(u)
	if err != nil {
		t.Error(err.Error())
	}

	if u.tr.GetID() == 0 {
		t.Error("Chiave non salvata")
	}

	tempName := *u.Name
	tempID := u.tr.GetID()

	err = record.LoadByID(u, u.GetTableRecord().GetID())
	if err != nil {
		t.Error(err.Error())
	}

	if tempName != *u.Name {
		t.Error("Campi non uguali")
	}

	u.Name = copy.String("Marco")

	err = record.Save(u)
	if err != nil {
		t.Error(err.Error())
	}

	if tempID != u.tr.GetID() {
		t.Error("Chiave primaria è cambiata durante l'update")
	}

	usersList, err := LoadAllUsers()
	if err != nil {
		t.Error(err.Error())
	}

	if len(usersList) == 0 {
		t.Error("Lunghezza inferiore al minimo nel contesto del test")
	}
}