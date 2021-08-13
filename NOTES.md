# Notes

## Actions in the logical plan

Usually broken in 4 categories:
- Index and Table Access
- Joins
- Sorting and Grouping
- Top-N Queries

Example from Postgres Query EXPLAIN
- SeqScan
- IndexScan
- IndexOnlyScan
- NestedLoop
- HashJoin
- MergeJoin
- GroupAggregate
- HashAggregate
- Sort / Sort Key
- Limit
- WindowAgg

Access and Filter predicates are also emitted:
- Access Predicates: (IndexCond) The access predicates express the start and stop conditions of the leaf node traversal.
- Index Filter Predicate: (IndexCond) Index filter predicates are applied during the leaf node traversal only. They do not contribute to the start and stop conditions and do not narrow the scanned range.
- Table level filter predicate: (Filter) Predicates on columns that are not part of the index are evaluated on the table level. For that to happen, the database must load the row from the heap table first.



If the Where condition is present in the index, Filter isn't used, instead an IndexCond is used.


Our Query Plan can include:
- Index and Table Access
  - TableScan
- Joins
  - NestedLoop
- Sorting and Grouping
  - Sort
- Top-N Queries
  -Limit
- Access and Filter:
  - Filter

### Test Cases plans:

Query: `SELECT a, b, c FROM t1 WHERE a = 1;`
Plan:
```
Filter{
     TableScan{
         Schema: csv
         RelationName: t1
     }
    Filter: (a = 1)
}
```

Query: `SELECT a, b, c FROM t1 WHERE a = 1 LIMIT 10;`
Plan:
```
Limit{
    Filter{
        TableScan{
            Schema: csv
            RelationName: t1
        }
        Filter: (a = 1)
    }
}
```

Query: `SELECT a, b, c FROM t1 WHERE a = 1 ORDER BY b LIMIT 10;`
Plan:
```
Limit{
    Sort{
        Filter{
            TableScan{
                Schema: csv
                RelationName: t1
            }
            Filter: "(a = 1)"
        }
        Key: b
        Method: default
    }
}
```

Query: `SELECT t1.a, t1.b, t1.c, t2.x FROM t1, t2 WHERE a = 1 AND t1.c = t2.y AND ORDER BY b LIMIT 10;`
Plan:
```
Limit{
    Sort{
        NestedLoop{
          JointType: "Left"
          Plans: []Plan{
              Filter{
                  ParentRelationship: OUTER
                  TableScan{
                      Schema: csv
                      RelationName: t1
                  }
              },
              Filter{
                  ParentRelationship: INNER
                  TableScan{
                      Schema: csv
                      RelationName: t1
                  }
              }
          }
        }
    }
}
```



