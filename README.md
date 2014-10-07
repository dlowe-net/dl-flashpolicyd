dl-flashpolicyd
===============

dl-flashpolicyd serves Flash policy files, which are used to grant
connection privileges to Flash applications.  If you want to use Flash
to connect to a host, you'll need something to serve the file.

The operation of this server is similar to that of the reference file
servers.  However, this server adds many more features.

* Can support more users by being much faster
* User privilege dropping (doesn't run as root)
* Configurable maximum policy file size
* Updating of policy file without rebooting
* Limited connection time to prevent DoS attacks

Building dl-flashpolicyd
------------------------

If you do not have a Go development workspace, you'll need to install
the Go programming environment and set up the workspace.

    $ apt-get install golang
    $ mkdir -p go/bin go/pkg go/src
    $ export GOPATH=go
    $ export PATH="$GOPATH/bin"

Now you can import this package into it:

   $ go get github.com/dlowe-net/dl-flashpolicyd
   $ go install github.com/dlowe-net/dl-flashpolicyd

Using dl-flashpolicyd
---------------------

First, create a flash policy file using the instructions on Adobe's
website.  The rest of this manual will assume you've read these
instructions.  Place the policy file in a location where you won't
forget about it.

