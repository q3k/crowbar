LIES LIES LIES
==============

THIS IS A *WIP* PIECE OF SOFTWARE, EVERYTHING UNDERNEATH IS A LIE, PLEASE MOVE ON.

Crowbar
=======

When a [corkscrew](http://www.agroman.net/corkscrew/) just isn't enough...

Intro
-----

Crowbar will tunnel TCP connections over ever a HTTP session using only GET and POST requests. This is in contrast to most tunneling systems that reuse the CONNECT verb. It provides basic authentication to make sure nobody steals your proxy to order drugs from Silkroad.

Features
--------

 - Establishes TCP connections
 - Authenticates users from an authentication file
 - Uses only basic HTTP verbs, so will probably work on most corporate firewalls

Security & Confidentiality
--------------------------

Crowbar *DOES NOT PROVIDE ANY DATA CONFIDENTIALITY*. While the user authentication mechanism protects from replay attacks to establish connectivity, it will not prevent someone from MITMing the later connection transfer itself. So, yeah, make sure to tunnel an SSH or OpenVPN over the tunnel.

The authentication code and crypto have not been reviewed by cryptographers. I am not a cryptographer. Just an FYI.

Is it any good?
---------------

Eh, it works. I'm not an experienced Golang programmer though, so the codebase is probably butt-ugly.

Building from source
--------------------

TODO because I'm a lazy fuck
