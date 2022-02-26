======================================
firebirdsql (Go firebird sql driver)
======================================

Firebird RDBMS https://firebirdsql.org SQL driver for Go

Requirements
-------------

* Firebird 2.5 or higher
* Golang 1.15 or higher

Example
-------------

::

   package main

   import (
       "fmt"
       "database/sql"
       _ "github.com/nakagami/firebirdsql"
   )

   func main() {
       var n int
       conn, _ := sql.Open("firebirdsql", "user:password@servername/foo/bar.fdb")
       defer conn.Close()
       conn.QueryRow("SELECT Count(*) FROM rdb$relations").Scan(&n)
       fmt.Println("Relations count=", n)

   }


See also driver_test.go

::

   package main

      import (
       "fmt"
       "github.com/nakagami/firebirdsql"
   )

   func main() {
       dsn := "user:password@servername/foo/bar.fdb"
       events := []string{"my_event", "order_created"}
       fbEvent, _ := firebirdsql.NewFBEvent(dsn)
       defer fbEvent.Close()
       sbr, _ := fbEvent.Subscribe(events, func(event firebirdsql.Event) { //or use SubscribeChan
           fmt.Printf("event: %s, count: %d, id: %d, remote id:%d \n", event.Name, event.Count, event.ID, event.RemoteID)
       })
       defer sbr.Unsubscribe()
       go func() {
               fbEvent.PostEvent(events[0])
               fbEvent.PostEvent(events[1])
       }()
       <- make(chan struct{}) //wait
   }

See also _example

Connection string
--------------------------

::

   user:password@servername[:port_number]/database_name_or_file[?params1=value1[&param2=value2]...]


General
=========

- user: login user
- password: login password
- servername: Firebird server's host name or IP address.
- port_number: Port number. default value is 3050.
- database_name_or_file: Database path (or alias name).

Optional
=========

param1, param2... are

.. csv-table::
   :header: Name,Description,Default,Note

   auth_plugin_name,Authentication plugin name.,Srp,Srp256/Srp/Legacy_Auth are available.
   column_name_to_lower,Force column name to lower,false,For "github.com/jmoiron/sqlx"
   role,Role name,
   timezone, Time Zone name,, For Firebird 4.0+
   wire_crypt,Enable wire data encryption or not.,true,For Firebird 3.0+
   charset, Firebird Charecter Set,
