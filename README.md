m3m3 is a simple web app allowing users to post comments (jokes) on current pieces of news, represented by a rolling list of images that is constantly fed by journals' RSS streams.

This is all work in progress. That's my toy project to learn Go.

If you want to try it, you will need to customize the file "config.example.toml" with your own settings (mysql server address/credentials + smtp + image folders ecc), then rename it to "config.toml". 
You can also customize the file "rss.conf", which contains the addresses of the rss feeds to be used by the crawler.

Then just compile m3m3.go and run the binary. 

Any comments/suggestions are of course welcome!!