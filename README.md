# ChittyChat
BSDISYS1KU-20222, mandatory handin 4

##  How to run

 1. Run `server.go` & `client.go` in separate terminals - the order is very important, the servers must be running first. 
 
 In the source code for `server.go` - you will find a `const BASEPORT`, this is the port that the servers will incrementally use. You do not need to supply a port through commandline, if the baseport is not available - a server will increment and try again.

 In the source code for `client.go` - you will find `const BASEPORT` also (this needs to match `server.go`), and `const REPLICAS`. While it is not necessary to set `const REPLICAS` to the same amount that of server instances you've started - it does make sense to do, since it prevents having to wait for timeouts to finish.

    ```console
    $ go run .\server\server.go
    $ go run .\server\server.go
    $ go run .\server\server.go
    $ go run .\server\server.go
    $ go run .\server\server.go
    $ go run .\client\client.go *
    ```
The client *needs* a parameter of uint32 - this is the ID of the client when bidding. There is nothing that checks for whether or not an ID is in use, so just make sure you do not use duplicates


##  Stuff that might go wrong
