m3m3 is a simple web app allowing users to post comments (jokes) on current pieces of news, represented by a rolling list of images that is constantly fed by journals' RSS streams.

A working version of it can be found at http://m3m3.ddns.net/

If you want to self host it, you will need to customize the file "config.example.toml" with your own settings (mysql server address/credentials + smtp + static files folder ecc), then rename it to "config.toml". You will need a mysql server running. You can also customize the file "rss.conf", which contains the addresses of the rss feeds to be used by the crawler. Then just compile m3m3.go and run the binary. 

Any comments/suggestions are of course welcome!!