Scoreserver
===========

[![Build Status](https://travis-ci.org/xyproto/scoreserver.svg?branch=master)](https://travis-ci.org/xyproto/scoreserver)

REST/JSON server for managing users and scores


Technologies involved
---------------------

* Server programming language: [go](https://golang.org)
* Authentication: [HTTP Basic Auth](https://en.wikipedia.org/wiki/Basic_access_authentication)
* Database backend: [Redis](http://redis.io/)
* API style: [REST](https://en.wikipedia.org/wiki/Representational_state_transfer) and [JSON](http://en.wikipedia.org/wiki/JSON)


Suggestions for additional technologies 
---------------------------------------

* Webserver for http/https, and proxy for server apps: [nginx](https://nginx.org)
* Encryption: [HTTP over TLS/SSL](http://en.wikipedia.org/wiki/HTTP_Secure)


Requirements
------------

* redis and go (the package is often named "golang")
* Use HTTPS whenever using HTTP Basic Auth


Terms used
----------

* ANY can be GET, POST, PUT or anything
* For an URL like /api/1.0/create/:username/:password, :username should be replaced with the desired username and :password with the desired password, when making the HTTP request.


Admin user management
---------------------

* HTTP GET **/**
  * Reveals if an administrator user has been created or not.

* HTTP ANY **/status**
  * Reveals if an administrator exists and the login status.

* HTTP GET **/admin**
  * Administration panel. Not implemented yet.

* HTTP GET **/register**
  * For registrating a new administrator by filling in a form.

* HTTP POST **/register**
  * For registering a new administrator, needs the form names 'password1', 'password2' and 'email'.
  * Username is 'admin' by default.
  * Only works if an administrator has not yet been registered.

* HTTP GET **/login**
  * For logging in the administrator by filling in a form.

* HTTP POST **/login**
  * For logging in the administrator, needs the form name 'password'.
  * Username is 'admin' by default.

* HTTP ANY **/logout**
  * For logging out the administrator.

TODO
----

* Administration panel
  * For changing the username for the administrator
  * For changing the password for the administrator
  * For listing and managing registered users


API calls
---------

The following calls requires authentication with HTTP Basic Auth, where the username is 'admin' and the password is set when creating the admin user with the /register call above.

* HTTP ANY **/api/1.0/**
  * Returns the JSON data: {"hello": "fjaselus"} as a test.

* HTTP POST **/api/1.0/create/:username**
  * Create a new user, with empty password and email.

* HTTP POST **/api/1.0/register/:username/:password/:email**
  * Create a new user, with a username, password and email address.

* HTTP POST **/api/1.0/login/:username/:password**
  * Log in a user, given a username and a password.

* HTTP ANY **/api/1.0/logout/:username**
  * Log out a user, given a username.

* HTTP ANY **/api/1.0/status/:username**
  * Show the login status for a given username.

* HTTP POST **/api/1.0/score/:username/:score**
  * Set a score for a given username.

* HTTP GET **/api/1.0/score/:username**
  * Return the score for a given username.


Port
----

Set the HOST and PORT environment variables for running the server on a different host/port. This is handled by the [martini](http://martini.codegangsta.io) package.


General information
-------------------

* Author: Alexander Rødseth, 2014
* License: MIT
