
Usher
=====

`usher` is a tiny personal url shortener that manages a simple local
database of `code => url` mappings, stored in a YAML text file.

Usher mappings can be published to various backends (e.g. Amazon S3)
to provide the actual redirection services.

Installation
------------

    go get github.com/gavincarr/usher


Usage
-----

    # Help output
    usher -h

    # Initialise a new database for your shortener domain
    usher init example.me

    # Add a mapping with an explicit code (both url+code and code+url work)
    usher add https://github.com/gavincarr/usher usher
    usher add github https://github.com/gavincarr/usher

    # Add a mapping with a randomly generated code
    usher add https://github.com/gavincarr/usher

    # List current mappings
    usher ls

    # Update an existing mapping to a new url
    usher update github https://github.com/gavincarr

    # Delete a mapping
    usher rm github

    # Report locations of config and database
    usher config
    usher db

    # Configure a backend to push to e.g. S3 e.g.
    $EDITOR $(usher config)

    # Push current mappings to your backend
    usher push



Author and Licence
------------------

Copyright 2020 Gavin Carr <gavin@openfusion.com.au>.

usher is available under the terms of the Apache 2.0 Licence.


