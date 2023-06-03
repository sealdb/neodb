/*
 * NeoDB
 *
 * Copyright 2019 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package shift

// Use flavor for different target cluster
const (
	ToMySQLFlavor   = "mysql"
	ToMariaDBFlavor = "mariadb"
	ToNeoDBFlavor   = "neodb"
)

type Config struct {
	ToFlavor string

	From         string
	FromUser     string
	FromPassword string
	FromDatabase string
	FromTable    string

	To         string
	ToUser     string
	ToPassword string
	ToDatabase string
	ToTable    string

	Rebalance              bool
	Cleanup                bool
	MySQLDump              string
	Threads                int
	Behinds                int
	NeoDBURL               string
	Checksum               bool
	WaitTimeBeforeChecksum int // seconds
}
