dl-flashpolicyd
===============

dl-flashpolicyd serves Flash policy files, which are used to grant
connection privileges to Flash applications.  The [Adobe website] (https://www.adobe.com/devnet/flashplayer/articles/socket_policy_files.html)
describes the protocol and format of the policy files.  If you want to use Flash
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

    apt-get install golang
    mkdir -p go/bin go/pkg go/src
    export GOPATH=go
    export PATH="$GOPATH/bin"

Now you can import this package into it:

    go get github.com/dlowe-net/dl-flashpolicyd
    go install github.com/dlowe-net/dl-flashpolicyd

Short Installation
------------------

If you are running Debian 7, run these commands:

    apt-get install daemon
    sudo -E $GOPATH/src/github.com/dlowe-net/dl-flashpolicyd/setup.sh <policy file>
    sudo service flashpolicyd start

Long Installation
-----------------

Long term, the best thing to do is to advocate for this program to be
packaged by your distribution.  If you're willing to forge ahead,
however, you can try these instructions.

You are strongly urged to create a new user especially for
dl-flashpolicyd, and not run it as root.

    sudo addgroup flashpolicyd
    sudo adduser --group flashpolicyd \
          --home /var/tmp \
          --no-create-home \
          --disabled-login \
          --ingroup nobody \
          flashpolicyd
    sudo mkdir -m=755 /var/log/flashpolicyd
    sudo chown flashpolicyd:flashpolicyd /var/log/flashpolicyd

Copy the executable from your Go workspace to a binary directory:

    sudo cp `which dl-flashpolicyd` /usr/sbin/flashpolicyd

dl-flashpolicyd doesn't actually daemonize itself, so you will need to
use a supervisor program to do so.  One such supervisor is called `daemon`.

    sudo apt-get/yum install daemon

You can run the flashpolicyd like this, but it won't automatically
restart on boot:

    sudo daemon -r -n dl-flashpolicyd \
          -D / \
          -o /var/log/flashpolicyd/flashpolicyd.log -- \
          /usr/sbin/flashpolicyd -user=flashpolicyd -file=<policy file>

The Linux world has many different ways of starting a daemon on boot,
and it can be hard to tell in advance what will work on which version.
Init scripts in /etc/init.d are likely to be supported for the near
future.

Feel free to email if you have problems.

Using dl-flashpolicyd
---------------------

First, create a flash policy file using the instructions on [Adobe's
website](https://www.adobe.com/devnet/flashplayer/articles/socket_policy_files.html).
The rest of this manual will assume you've read these instructions.
You can also read the [formal specification of the policy file]
(https://www.adobe.com/devnet/articles/crossdomain_policy_file_spec.html),
which also includes DTDs for XML validation.
Place the policy file in a location where you won't forget about it.

You'll probably need to test it out.  Run on the command line:

    sudo $GOPATH/bin/dl-flashpolicyd -update=1s -file=<policy file>

After some diagnostics, you should be able to try your app and see if
it connects properly to the host.  You should also see a log message
printed out by dl-flashpolicyd that the policy file was served.  You
can edit and try connecting again, since it will check for a new
version of the file every second.  Press control-C to kill the
program.

Once you have a policy file that does what you want, you will need to
run it with production settings and ensure that it comes up when the
host is rebooted.

If you were able to follow the installation procedure, starting the
policy server is fairly simple:

    sudo service flashpolicyd start

The default installation will save its logs in /var/log/flashpolicyd.
You should check there to make sure that the policy server is
functioning properly.
