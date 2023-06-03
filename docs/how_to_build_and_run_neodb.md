`NeoDB` is fairly simple to deploy, without installing external dependencies.

---

# Contents

- [Contents](#contents)
- [How to build and run neodb](#how-to-build-and-run-neodb)
  - [Requirements](#requirements)
  - [Step1. Download src code from github](#step1-download-src-code-from-github)
  - [Step2. Build](#step2-build)
  - [Step3. Run neodb](#step3-run-neodb)
  - [Step4. Add a backend(mysql server) to neodb](#step4-add-a-backendmysql-server-to-neodb)
  - [Step5. Connect mysql client to neodb](#step5-connect-mysql-client-to-neodb)

# How to build and run neodb

## Requirements

1. [Go](http://golang.org) version 1.8 or newer is required.("sudo apt install golang" for ubuntu or "yum install golang" for centOS/redhat)
2. A 64-bit system is strongly recommended. Building or running neodb on 32-bit systems has not been tested

## Step1. Download src code from github

```
$ git clone https://github.com/sealdb/neodb
```

## Step2. Build

After download neodb src code from github, it will generate a directory named "neodb", execute the following commands:

```
$ cd neodb
$ make build
```

The binary executable file is in the "bin" directory, execute command "ls bin/":

```
$ ls bin/

---Response---
$ neodb neodbcli
```

## Step3. Run neodb

Copy the default configure file conf/neodb.default.json into bin first:

```
$ cp conf/neodb.default.json bin/
```

Then run `neodb` server:

```
$ bin/neodb -c bin/neodb.default.json
```

If start successfully, you will see infos next:

```
    neodb:[{Tag:rc-20180126-16-gf448be1 Time:2018/04/04 03:31:39 Git:f448be1
    GoVersion:go1.8.3 Platform:linux amd64}]
    2018/04/04 15:20:17.136839 proxy.go:79:
     ....
     ....
    2018/04/04 15:20:17.151499 admin.go:54:      [INFO]     http.server.start[:8080]...
```

When neodb started, it will use three ports:
`3308: External service port for MySQL client link`
`8080: Management port, external RESTFUL interface`
`6060: debug port, golang debug port`

## Step4. Add a backend(mysql server) to neodb

This is an admin instruction of neodb api, for more admin instructions, see [neodb admin API](api.md).

First, create an account on the MySQL server, and then add the MySQL server as a backend to neodb by using the account. NeoDB uses the account to connect to the backend.

Here we suppose mysql has being installed and the mysql service has beeing started on your machine, the user and password logged in to mysql are all root.

`user`: the user to login mysql
`password`: the password to login mysql

```
$ curl -i -H 'Content-Type: application/json' -X POST -d \
> '{"name": "backend1", "address": "127.0.0.1:3306", "user":\
>  "root", "password": "root", "max-connections":1024}' \
> http://127.0.0.1:8080/v1/neodb/backend
```

`Response: `

```
HTTP/1.1 200 OK
Date: Mon, 09 Apr 2018 03:23:02 GMT
Content-Length: 0
Content-Type: text/plain; charset=utf-8
```

The backends information is recorded in the JSON file `$meta-dir\backend.json`.

```
{
        "backends": [
                {
                        "name": "backend1",
                        "address": "127.0.0.1:3306",
                        "user": "root",
                        "password": "root",
                        "database": "",
                        "charset": "utf8",
                        "max-connections": 1024
                }
        ]
}
```

## Step5. Connect mysql client to neodb

NeoDB supports client connections to the MySQL protocol, like: mysql -uroot -h127.0.0.1 -P3308
`root`:account login to neodb, we provide default account 'root' with no password to login
`3308`:neodb default port

```
$ mysql -uroot -h127.0.0.1 -P3308
```

If connected success, you will see:

```
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 1
Server version: 5.7-NeoDB-1.0

Copyright (c) 2000, 2018, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql>
```

Now you can send sql from mysql client, for more sql supported by neodb sql protocol, see [neodb_sql_statements_manual](neodb_sql_statements_manual.md)
`Example: `

```
mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| db_gry_test        |
| db_test1           |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
6 rows in set (0.01 sec)
```
