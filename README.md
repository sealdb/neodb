[![Github Actions Status](https://github.com/sealdb/neodb/workflows/NeoDB%20Build/badge.svg?event=push)](https://github.com/sealdb/neodb/actions?query=workflow%3A%NeoDB+Build%22+event%3Apush)
[![Github Actions Status](https://github.com/sealdb/neodb/workflows/NeoDB%20Test/badge.svg?event=push)](https://github.com/sealdb/neodb/actions?query=workflow%3A%NeoDB+Test%22+event%3Apush)
[![Github Actions Status](https://github.com/sealdb/neodb/workflows/NeoDB%20Coverage/badge.svg?event=push)](https://github.com/sealdb/neodb/actions?query=workflow%3A%NeoDB+Coverage%22+event%3Apush)
[![Go Report Card](https://goreportcard.com/badge/github.com/sealdb/neodb)](https://goreportcard.com/report/github.com/sealdb/neodb)
[![codecov.io](https://codecov.io/gh/sealdb/neodb/graphs/badge.svg)](https://codecov.io/gh/sealdb/neodb/branch/main)

# OverView

NeoDB is an open source, Cloud-native MySQL database for unlimited scalability and performance.

## What is NeoDB?

NeoDB is a cloud-native database based on MySQL, and architected in fully distributed cluster that enable unlimited scalability (scale-out), capacity and performance. It supported distributed transaction that ensure high data consistency, and leveraged MySQL as storage engine for trusted data reliability. NeoDB is compatible with MySQL protocol, and sup-porting automatic table sharding as well as batch of automation feature for simplifying the maintenance and operation workflow.

## Features

- **Automatic Sharding**
- **Auditing and Logging**
- **Parallel Execution**: Parallel Query, Parallel DML and Parallel DDL
- **Parallel CHECKSUM TABLE**: Gives same results as MySQL
- **Distributed Transaction**: Snapshot Isolation
- **Distributed Joins**: Sort-Merge Join, Nested-Loop Join
- **Distributed Full Text Search**
- **Multi Tenant by Database**
- **Prepared SQL Statement**
- **JSON**

## Documentation

For guidance on installation, deployment, and administration, see our [Documentation](docs).

## Architecture

## Overview

NeoDB is a new generation of distributed relational database (MyNewSQL) based on MySQL. It was designed to create the open-source database our developers would want to use: one that has features such as financial high availability、
large-capacity database、automatic plane split table、 scalable and strong consistency, this guide sets out to detail the inner-workings of the neodb process as a means of explanation.

## SQL Layer

### SQL support

On SQL syntax level, NeoDB Fully compatible with MySQL.You can view all of the SQL features NeoDB supports here [neodb_sql_statements_manual](docs/neodb_sql_statements_manual.md)

### SQL parser, planner, executor

After your SQL node receives a SQL request from a mysql client via proxy, NeoDB parses the statement, creates a query plan, and then executes the plan.

                                                                    +---------------+
                                                        x---------->|node1_Executor |
                                +--------------------+  x           +---------------+
                                |      SQL Node      |  x
                                |--------------------|  x
    +-------------+             |     sqlparser      |  x           +---------------+
    |    query    |+----------->|                    |--x---------->|node2_Executor |
    +-------------+             |  Distributed Plan  |  x           +---------------+
                                |                    |  x
                                +--------------------+  x
                                                        x           +---------------+
                                                        x---------->|node3_Executor |
                                                                    +---------------+

`Parsing`

Received queries are parsed by sqlparser (which describes the supported syntax by mysql) and generated Abstract Syntax Trees (AST).

`Planning`

With the AST, NeoDB begins planning the query's execution by generating a tree of planNodes.
This step also includes steps analyzing the client's SQL statements against the expected AST expressions, which include things like type checking.

You can see the a query plan generates using `EXPLAIN`(At this stage we only use `EXPLAIN` to analysis Table distribution).

`Excuting`

Executing an Executor in a storage layer in Parallel with a Distributed Execution Plan.

### SQL with Transaction

The SQL node is stateless, but in order to guarantee transaction `Snapshot Isolation`, it is currently a write-multiple-read mode.

## Transaction Layer

` Distributed transaction`

NeoDB provides distributed transaction capabilities. If the distributed executor at different storage nodes and one of the nodes failed to execute, then operation of the rest nodes will be rolled back, This guarantees the atomicity of operating across nodes and makes the database in a consistent state.

`Isolation Levels`

NeoDB achieves the level of SI (Snapshot Isolation) at the level of consistency. As long as a distributed transaction has not commit, or if some of the partitions have committed, the operation is invisible to other transactions.

` Transaction with SQL Layer`

The SQL node is stateless, but in order to guarantee transaction `Snapshot Isolation`, it is currently a write-multiple-read mode.

## Issues

The [integrated github issue tracker](https://github.com/sealdb/neodb/issues)
is used for this project.

## License

NeoDB is released under the GPLv3. See LICENSE
