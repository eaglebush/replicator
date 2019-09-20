package replicator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/eaglebush/datahelper"
)

// Replicator - A replicator for message-given data
type Replicator struct {
	DecimalSign    rune
	DigitSeparator rune
	Debug          bool
	DropReplicated bool
	Subjects       []Item
}

// Column - replication column for initialization of replicator
type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Null *bool  `json:"null,omitempty"`
}

// Item - an item in the replicator
type Item struct {
	SubjectRoot string
	Columns     []Column
	DataKeys    []string
}

var subjectTable map[string]string
var dataKeys map[string][]string

// NewReplicator - create a new replicator with defaults
func NewReplicator(debug bool) *Replicator {
	return &Replicator{
		DecimalSign:    '.',
		DigitSeparator: ',',
		Debug:          debug,
	}
}

// Init - initialize table
func (r *Replicator) Init(dh *datahelper.DataHelper, subject string, tableColumns []Column, dataKeyColumns []string) error {
	tableName, err := BuildTableName(subject)
	if err != nil {
		return err
	}

	// Check if table is present in the subjectTable map
	for k := range subjectTable {
		if k == tableName {
			return nil
		}
	}

	for _, rep := range []string{`.`, `,`} {
		tableName = strings.Replace(tableName, rep, `_`, -1)
	}

	if subjectTable == nil {
		subjectTable = make(map[string]string)
		dataKeys = make(map[string][]string)
	}
	subjectTable[tableName] = tableName
	dataKeys[tableName] = dataKeyColumns

	if r.DropReplicated {
		r.Drop(dh, subject)
	}

	tr := true

	cnt := len(tableColumns) - 1
	sql := ``
	sql += `CREATE TABLE ` + tableName + ` (`
	for i, v := range tableColumns {
		sql += `	` + v.Name + ` ` + v.Type + ` `

		if v.Null == nil {
			v.Null = &tr
		}

		if !*v.Null {
			sql += `NOT `
		}
		sql += `NULL`

		if cnt != i {
			sql += `,`
		}
		sql += "\r\n"
	}
	sql += `);`

	if r.Debug {
		log.Println(sql)
	}

	// create table it does not exist
	_, err = dh.Exec(sql)
	if err != nil {
		// table might have existed
		log.Printf("Error Init: %s", err.Error())
	}

	return nil
}

// Insert - insert action to database
func (r *Replicator) Insert(dh *datahelper.DataHelper, subject string, msgData []byte) error {
	tableName, err := BuildTableName(subject)
	if err != nil {
		return err
	}

	//var objmap map[string]interface{}
	var objmap map[string]json.RawMessage
	err = json.Unmarshal(msgData, &objmap)
	if err != nil {
		return err
	}

	cma := ``
	cols := ``
	vals := ``
	sql := ``

	for k, v := range objmap {
		cols += cma + k

		vals += cma + `'` + strings.Replace(r.rawtstr(v), `'`, `''`, -1) + `'`
		cma = `, `
	}

	tbl := subjectTable[tableName]
	if len(tbl) == 0 {
		return errors.New(`The table name could not be found from the given subject`)
	}

	sql += `INSERT INTO ` + tbl + ` (` + cols + `) VALUES (` + vals + `);`

	if r.Debug {
		log.Println(sql)
	}

	_, err = dh.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

// Update - update action to database
func (r *Replicator) Update(dh *datahelper.DataHelper, subject string, msgData []byte) error {
	tableName, err := BuildTableName(subject)
	if err != nil {
		return err
	}

	var objmap map[string]json.RawMessage
	err = json.Unmarshal(msgData, &objmap)

	cma := ``
	colvals := ``
	sql := ``
	valid := true

	var dataKeyValues map[string][]byte
	dataKeyValues = make(map[string][]byte)

	for k, v := range objmap {
		valid = true

		// check if the column is in the datakeys
		for _, f := range dataKeys[tableName] {
			if f == k {
				dataKeyValues[f] = v
				valid = false
			}
		}

		if valid {
			colvals += cma + k + `='` + strings.Replace(r.rawtstr(v), `'`, `''`, -1) + `'`
			cma = `, `
		}

	}

	tbl := subjectTable[tableName]
	if len(tbl) == 0 {
		return errors.New(`The table name could not be found from the given subject`)
	}

	sql += `UPDATE ` + tbl + ` SET ` + colvals + ` WHERE `

	// Get filters
	cma = ``
	for k, v := range dataKeyValues {
		sql += cma + k + `='` + strings.Replace(r.rawtstr(v), `'`, `''`, -1) + `'`
		cma = ` AND `
	}

	if r.Debug {
		log.Println(sql)
	}

	_, err = dh.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

// Delete - delete a record in the database
func (r *Replicator) Delete(dh *datahelper.DataHelper, subject string, msgData []byte) error {
	tableName, err := BuildTableName(subject)
	if err != nil {
		return err
	}

	var objmap map[string]json.RawMessage
	err = json.Unmarshal(msgData, &objmap)

	cma := ``
	sql := ``

	var dataKeyValues map[string][]byte
	dataKeyValues = make(map[string][]byte)

	for k, v := range objmap {
		// check if the column is in the datakeys
		for _, f := range dataKeys[tableName] {
			if f == k {
				dataKeyValues[f] = v
			}
		}
	}

	sql += `DELETE FROM ` + subjectTable[tableName] + ` WHERE `

	// Get filters
	cma = ``
	for k, v := range dataKeyValues {
		sql += cma + k + `='` + strings.Replace(r.rawtstr(v), `'`, `''`, -1) + `'`
		cma = ` AND `
	}

	if r.Debug {
		log.Println(sql)
	}

	_, err = dh.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

// Drop - drop replicated table
func (r *Replicator) Drop(dh *datahelper.DataHelper, subject string) error {
	tableName, err := BuildTableName(subject)
	if err != nil {
		return err
	}

	sql := `DROP TABLE ` + tableName + `;`

	if r.Debug {
		log.Println(sql)
	}

	_, err = dh.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

// LoadReplicator - load replicator table definitions from configuration file
func LoadReplicator(dh *datahelper.DataHelper, replicatorConfig string, dropReplicatedTables bool) (*Replicator, error) {
	b, err := ioutil.ReadFile(replicatorConfig)
	if err != nil {
		return nil, err
	}

	var rp Replicator
	err = json.Unmarshal(b, &rp)
	if err != nil {
		return nil, err
	}

	for _, v := range rp.Subjects {
		err = rp.Init(dh, v.SubjectRoot, v.Columns, v.DataKeys)
		if err != nil {
			return nil, err
		}
	}

	return &rp, nil
}

func (r *Replicator) rawtstr(value []byte) string {
	var b string

	bstr := strings.Trim(string(value), `"`)
	var err error
	var f interface{}

	// check if the string is a date
	f, err = time.Parse(time.RFC3339Nano, bstr)
	if err == nil {
		b = f.(time.Time).Format(`2006-01-02 15:04:05`)
		return b
	}

	// check if the string is a date
	f, err = time.Parse(time.RFC3339, bstr)
	if err == nil {
		b = f.(time.Time).Format(`2006-01-02 15:04:05`)
		return b
	}

	f, err = strconv.ParseUint(bstr, 10, 64)
	if err == nil {
		b = strconv.FormatUint(f.(uint64), 10)
		return b
	}

	f, err = strconv.ParseInt(bstr, 10, 64)
	if err == nil {
		b = strconv.FormatInt(f.(int64), 10)
		return b
	}

	f, err = strconv.ParseFloat(bstr, 32)
	if err == nil {
		// check the length of the decimal places
		dotp := strings.Index(bstr, string(r.DecimalSign))
		numdec := `2`
		if dotp != -1 {
			numdec = strconv.Itoa(len(bstr[dotp+1:]))
		}
		ffmt := "%." + numdec + "f"

		b = fmt.Sprintf(ffmt, float32(f.(float64)))
		return b
	}

	f, err = strconv.ParseFloat(bstr, 64)
	if err == nil {
		// check the length of the decimal places
		dotp := strings.Index(bstr, string(r.DecimalSign))
		numdec := `2`
		if dotp != -1 {
			numdec = strconv.Itoa(len(bstr[dotp+1:]))
		}
		ffmt := "%." + numdec + "f"

		b = fmt.Sprintf(ffmt, f.(float64))
		return b
	}

	f, err = strconv.ParseBool(bstr)
	if err == nil {
		b = "0"
		s := strings.ToLower(bstr)
		if len(s) > 0 {
			if s == "true" || s == "on" || s == "yes" || s == "1" || s == "-1" || s == "t" {
				b = "1"
			}
		}
		return b
	}

	b = bstr
	return b
}

// func anytstr(value interface{}) string {
// 	var b string
// 	const longForm = `2006-01-02 15:04:05`

// 	switch value.(type) {
// 	case string:
// 		bstr := value.(string)

// 		// check if the string is a date
// 		t, err := time.Parse(time.RFC3339, bstr)
// 		if err == nil {
// 			b = t.Format(longForm)
// 			break
// 		}

// 		b = bstr
// 	case int:
// 		b = strconv.FormatInt(int64(value.(int)), 10)
// 	case int8:
// 		b = strconv.FormatInt(int64(value.(int8)), 10)
// 	case int16:
// 		b = strconv.FormatInt(int64(value.(int16)), 10)
// 	case int32:
// 		b = strconv.FormatInt(int64(value.(int32)), 10)
// 	case int64:
// 		b = strconv.FormatInt(value.(int64), 10)
// 	case uint:
// 		b = strconv.FormatUint(uint64(value.(uint)), 10)
// 	case uint8:
// 		b = strconv.FormatUint(uint64(value.(uint8)), 10)
// 	case uint16:
// 		b = strconv.FormatUint(uint64(value.(uint16)), 10)
// 	case uint32:
// 		b = strconv.FormatUint(uint64(value.(uint32)), 10)
// 	case uint64:
// 		b = strconv.FormatUint(uint64(value.(uint64)), 10)
// 	case float32:
// 		b = fmt.Sprintf("%f", value.(float32))
// 	case float64:
// 		b = fmt.Sprintf("%f", value.(float64))
// 	case bool:
// 		b = "0"
// 		s := strings.ToLower(value.(string))
// 		if len(s) > 0 {
// 			if s == "true" || s == "on" || s == "yes" || s == "1" || s == "-1" {
// 				b = "1"
// 			}
// 		}
// 	case time.Time:
// 		b = value.(time.Time).Format(longForm)
// 	}

// 	return b
// }

// BuildTableName - Build replicator table
func BuildTableName(subject string) (string, error) {
	subject = strings.TrimSpace(strings.ToLower(subject))

	// Check subject composition. Default composition should be in <api>.<table>.<event> format. There must be 2 dots in the subject
	tbln := strings.Split(subject, `.`)
	if len(tbln) != 3 {
		return ``, errors.New(`The subject is not compliant with the name convention. Convention should be in <api>.<table>.<event> format`)
	}

	// create a table name from the event/subject
	tableName := tbln[0] + `.` + tbln[1]

	for _, rep := range []string{`.`, `,`} {
		tableName = strings.Replace(tableName, rep, `_`, -1)
	}

	return tableName, nil
}
