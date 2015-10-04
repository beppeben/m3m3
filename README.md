m3m3 is (will be) a web app allowing users to post comments on current pieces of news, represented by a rolling list of images that is constantly fed by journals' RSS streams.

The repository contains backend Go code providing webservices to be used by frontends. It is structured into different modules:
- crawler: contains methods to retrieve images from rss and insert them into a rolling data structure (where items will sorted by date/likes/n_comments, in a way yet to be clarified)
- db: manages connections with the database (info on db location/credentials in a separate config file)
- server: provides basic API to authenticate, retrieve news, post comments, ecc..

This is all work in progress. That's my toy project to learn Go.

To try it, just compile m3m3.go and run the binary. You will need a config file, I'll soon post a template.
