Crowbar
=======

When a [corkscrew](http://www.agroman.net/corkscrew/) just isn't enough...

Intro
-----

![Crowbar overview](http://q3k.org/crowbar-overview.png)

Crowbar is a tool that allows you to establish secure VPN with your existing TCP endpoints (an OpenVPN setup, an SSH for forwarding...) when your network connection is limited by a Web proxy that only allows basic port 80 HTTP connectivity.

Crowbar will tunnel TCP connections over ever a HTTP session using only GET and POST requests. This is in contrast to most tunneling systems that reuse the CONNECT verb. It also provides basic authentication to make sure nobody who stumbles upon the server steals your proxy to order drugs from Silkroad.

Features
--------

 - Establishes TCP connections via a proxy server using only HTTP GET and POST requests
 - Authenticates users from an authentication file
 - Will probably get you fired if you use this in an office setting

Security & Confidentiality
--------------------------

Crowbar **DOES NOT PROVIDE ANY DATA CONFIDENTIALITY**. While the user authentication mechanism protects from replay attacks to establish connectivity, it will not prevent someone from MITMing the later connection transfer itself, or fror MITMing whole sessions. So, yeah, make sure to tunnel an SSH or OpenVPN over the tunnel, and firewall off most outgoing connections on your proxy server (ie. only allow access to an already publically-available SSH server)

The authentication code and crypto have not been reviewed by cryptographers. I am not a cryptographer. You should consider this when deploying Crowbar.

Known bugs
----------

The crypto can be improved vastly to enable server authentication and make MITMing more difficult. It could also use a better authentication setup to allow the server to keep password hashes instead of plaintext.

The server lacks any cleanup functions and rate limiting, so it will leak both descriptors and memory - this should be fixed soon.

Is it any good?
---------------

Eh, it works. I'm not an experienced Golang programmer though, so the codebase is probably butt-ugly.

Usage
=====

Server setup
------------

This assumes you're using Linux. If not, you're on your own.

Set up an user for the service

    useradd -rm crowbard
    mkdir /etc/crowbar/
    chown crowbar:crowbar /etc/crowbar

Create an authentication file - a new-line delimited file containing username:password pairs.

    touch /etc/crowbar/userfile
    chown crowbar:crowbar /etc/crowbar/userfile
    chmod 600 /etc/crowbar/userfile
    echo -ne "q3k:supersecurepassword\n1337h4xx0r:canttouchthis" >> /etc/crowbar/userfile

Set up an iptables rule to forward traffic from the :80 port to :8080, where the server will be running. Replace eth0 with your public network interface.

    iptables -t nat -A PREROUTING -i eth0 -p tcp --dport 80 -j DNAT --to-port 8080

Run the daemon in screen/tmux or write some unit files for your distribution:

    crowbard -userfile=/etc/crowbard/userfile

Client setup
------------

This assumes you're running Linux on your personal computer. If not, you're on your own.

Crowbar will honor the _de-facto_ standard HTTP\_PROXY env var on Linux:

    export HTTP_PROXY=evil.company.proxy.com:80

For netcat-like functionality:

    crowbar-forward -local=- -username q3k -server http://your.proxy.server.com:80 -remote towel.blinkenlights.nl:23

For port-forwarding:


    crowbar-forward -local=127.0.0.1:1337 -username q3k -server http://your.proxy.server.com:80 -remote towel.blinkenlights.nl:23 &
    nc 127.0.0.1 1337


For SSH ProxyCommand integration, place this in your .ssh/config, and then SSH into your.ssh.host.com as usual:

    Host your.ssh.host.com
        ProxyCommand crowbar-forward -local=- -username q3k -password secret -server http://your.proxy.server.com:80 -remote %h:%p 

Building from source
--------------------

I assume you have a working $GOPATH.

    go get github.com/tools/godep
    go get github.com/q3k/crowbar
    godep restore github.com/q3k/crowbar/...
    go install github.com/q3k/crowbar/...

