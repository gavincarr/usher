
Usher
=====

Usher is a tiny personal Url SHortenER that manages a simple local
database of `code => url` mappings, stored in a YAML text file.

Usher mappings can be published to various backends (currently
Amazon S3 and render.com) to provide the actual redirection services.


Installation
------------

    go get github.com/gavincarr/usher/cmd/usher


Usage
-----

### Initialise a new database for the domain you want to use

    usher init example.me

### Add, list, update, and remove mappings

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

### Configure and publish to desired backend

    # Report locations of usher root directory, config and database
    usher root
    usher config
    usher db

    # Configure a backend to push to (`type: s3` or `type: render`) e.g.
    $EDITOR $(usher config)

    # Push current mappings to your backend
    usher push

### Help

    usher -h


Author and Licence
------------------

Copyright 2020 Gavin Carr <gavin@openfusion.com.au>.

usher is available under the terms of the MIT Licence.

