package gemini

import (
	"fmt"
	"math/rand"
	"strings"
)

type Keyspace struct {
	Name string
}

type ColumnDef struct {
	Name string
	Type string
}

type Table struct {
	Name           string
	PartitionKeys  []ColumnDef
	ClusteringKeys []ColumnDef
	Columns        []ColumnDef
}

type Schema interface {
	GetDropSchema() []string
	GetCreateSchema() []string
	GenMutateOp() string
	GenCheckOp() string
}

type schema struct {
	keyspace Keyspace
	table    Table
}

func (s *schema) GetDropSchema() []string {
	return []string{
		fmt.Sprintf("DROP KEYSPACE IF EXISTS %s", s.keyspace.Name),
	}
}

func (s *schema) GetCreateSchema() []string {
	createKeyspace := fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1}", s.keyspace.Name)
	partitionKeys := []string{}
	clusteringKeys := []string{}
	columns := []string{}
	for _, pk := range s.table.PartitionKeys {
		partitionKeys = append(partitionKeys, pk.Name)
		columns = append(columns, fmt.Sprintf("%s %s", pk.Name, pk.Type))
	}
	for _, ck := range s.table.ClusteringKeys {
		clusteringKeys = append(clusteringKeys, ck.Name)
		columns = append(columns, fmt.Sprintf("%s %s", ck.Name, ck.Type))
	}
	for _, cdef := range s.table.Columns {
		columns = append(columns, fmt.Sprintf("%s %s", cdef.Name, cdef.Type))
	}
	createTable := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (%s, PRIMARY KEY ((%s), %s))", s.keyspace.Name, s.table.Name, strings.Join(columns, ","), strings.Join(partitionKeys, ","), strings.Join(clusteringKeys, ","))
	return []string{
		createKeyspace,
		createTable,
	}
}

func (s *schema) GenMutateOp() string {
	columns := []string{}
	values := []string{}
	for _, pk := range s.table.PartitionKeys {
		columns = append(columns, pk.Name)
		values = append(values, fmt.Sprintf("%d", rand.Intn(100)))
	}
	for _, pk := range s.table.ClusteringKeys {
		columns = append(columns, pk.Name)
		values = append(values, fmt.Sprintf("%d", rand.Intn(100)))
	}
	for _, cdef := range s.table.Columns {
		columns = append(columns, cdef.Name)
		values = append(values, fmt.Sprintf("%d", rand.Intn(100)))
	}
	return fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s)", s.keyspace.Name, s.table.Name, strings.Join(columns, ","), strings.Join(values, ","))
}

func (s *schema) GenCheckOp() string {
	query := fmt.Sprintf("SELECT * FROM %s.%s", s.keyspace.Name, s.table.Name)
	if rand.Intn(2) == 1 {
		query += fmt.Sprintf(" ORDER BY %s", s.table.Columns[0].Name)
		if rand.Intn(2) == 1 {
			query += " ASC"
		}
	}
	if rand.Intn(2) == 1 {
		query += fmt.Sprintf(" LIMIT %d", rand.Intn(100))
	}
	return query
}

type SchemaBuilder interface {
	Keyspace(Keyspace) SchemaBuilder
	Table(Table) SchemaBuilder
	Build() Schema
}

type schemaBuilder struct {
	keyspace Keyspace
	table    Table
}

func (s *schemaBuilder) Keyspace(keyspace Keyspace) SchemaBuilder {
	s.keyspace = keyspace
	return s
}

func (s *schemaBuilder) Table(table Table) SchemaBuilder {
	s.table = table
	return s
}

func (s *schemaBuilder) Build() Schema {
	return &schema{keyspace: s.keyspace, table: s.table}
}

func NewSchemaBuilder() SchemaBuilder {
	return &schemaBuilder{}
}
