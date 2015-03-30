[chronicle-server](https://github.com/korylprince/chronicle-server)

This server has been split from the [old repo](https://github.com/korylprince/chronicle), and is completely refactored.

The client may be found [here](https://github.com/korylprince/chronicle-client).

chronicle is for simple client/server data logging to a database.
The client will send the following data:
 
* uid of current user
* username of current user
* full name (GECOS) of current user
* serial number of computer
* [munki](http://munki.github.io/munki/) ClientIdentifier
* hostname of computer
* local IP address of computer

The server logs this to a database along with the time and IP of the client as seen by the server.

The client will only work on OS X, and has only been tested on 10.9. The server will probably run anywhere.

#Installing#

`go get github.com/korylprince/chronicle-server`

`sql/v1.1.1.sql` will create the table and indexes. Make sure the database uses a utf8 collation, or convert the tables to utf8 after creating them with `ALTER TABLE <tablename> CONVERT TO CHARACTER SET utf8;`

`github.com/korylprince/chronicle-server/util` is a commandline tool that will migrate data from the old v1.1 schema.

If you have any issues or questions, email the email address below, or open an issue at:
https://github.com/korylprince/chronicle-server/issues

#Usage#

Read the source. It's pretty simple and readable.

The following Enviroment Variables are configurable:

* Server:

    * CHRONICLE_SQLDRIVER string //required
    * CHRONICLE_SQLDSN    string //required

    * CHRONICLE_WORKERS       int //default: 10
    * CHRONICLE_WRITEINTERVAL int //in seconds; default:15s

    * CHRONICLE_LISTENADDR string //addr format used for net.Dial; required
    * CHRONICLE_PREFIX     string //url prefix to mount api to without trailing slash

#Copyright Information#

Copyright 2015 Kory Prince (korylprince at gmail dot com.)

This code is licensed under the same license go is licensed under with slight modification (my name in place of theirs.) If you'd like another license please email me.
