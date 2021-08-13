package sql

type DataType int

const (
	NULL DataType = iota
	TEXT
	REAL
	INTEGER
	DATETIME
	BOOLEAN
	BLOB
)

var dataTypes = map[DataType]string{
	NULL:     "NULL",
	TEXT:     "TEXT",
	REAL:     "REAL",
	INTEGER:  "INTEGER",
	DATETIME: "DATETIME",
	BOOLEAN:  "BOOLEAN",
	BLOB:     "BLOB",
}

func (d DataType) String() string {
	if str, ok := dataTypes[d]; ok {
		return str
	}
	return "N/A"
}
