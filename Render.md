
Usher Setup on Render
=====================

Render (render.com) is a newish PAAS startup that offers free hosting
for static websites, with built-in HTTPS and CDN support, as well as
some nice paid application hosting options.

For usher, render is a much easier backend to configure than Amazon S3,
plus you get built-in HTTPS and CDN support, all for free. (Of course,
it would be great to support them by using them for paid hosting too, if
you're able.)


Usher Configuration
-------------------

On the usher side, all that's required to configure a render backend is
to set `type: render` in your usher config file e.g.

    $ cat $(usher config)
    example.me:
      type: render

    $ usher push

When you do your `usher push` usher will just create a local `render.yaml`
file in your usher root directory instead of pushing the data somewhere
remote.

To actually deploy the data remotely we use git instead of usher, by
setting up our usher root directory as a git repo and using github or
gitlab as our remote.

For example, if you've created an empty github repo called e.g.
`usher_example`, you might do:

     $ cd $(usher root)
     $ git init
     $ git add .
     $ git commit -m 'Initial import'
     $ git remote add origin git@github.com:USERNAME/usher_example
     $ git push origin master

and your usher database and the `render.yaml` file should show up in your
github repo.


Render Configuration
--------------------

On the render side, you need to:

1. Signup for a free account at https://render.com/

2. Login to render, and under "Account Settings" connect your account to
   github or gitlab. It's fine to select the just connect your usher repos,
   at least initially.

3. Click the "YAML" link in the left-hand menu, then click the
   "New From YAML" button, then select the usher repo you created above.

4. Once selected, you’ll see a list of the changes that will be applied
   based on the contents of the usher render.yaml file. If everything looks
   good, click on Apply Changes to create the resources.

5. Render should then generate a new deploy, which if successful will have
   created your static redirects. You should see a link near the top of the
   page with an `onrender.com` domain like `https://example-me.onrender.com`,
   which is your base url.

   You can then do some tests by appending the codes you've configured to
   that url e.g.

      https://example-me.onrender.com/test1
      https://example-me.onrender.com/test2

      # And if you've created an 'INDEX' code:
      https://example-me.onrender.com/


DNS Configuration
-----------------

To configure your domain to point to render, do:

1. From your render dashboard select your domain website, then "Settings",
   and then click the "Add Custom Domain" button.

2. Add your domain (without a `www`) in the text box and hit "Save".

3. Follow the DNS instructions that come up with your DNS provider. This
   involves adding an `ANAME` or `ALIAS` DNS record from your domain to
   the render hosting domain e.g. `example.me => example-me.onrender.com`,
   and a similar CNAME for `www.$DOMAIN`.

   If your DNS provider doesn't support ANAME/ALIAS records, you can add
   an A record to the given ip address instead.

If all goes well Render should report your domains as "Verified", and then
"Certificate Issued", at which point you should be able to query your
domain urls (in both https and http variants) and see the redirects working
e.g.

    https://example.me/test1
    https://example.me/test2
    http://example.me/test1
    http://example.me/test2

