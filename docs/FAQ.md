# Frequently Asked Questions

### Q: Why the name change from telegraf to telex

Telex is a direct copy of Telegraf but scaled down by removing and consolidating functionality.
It also uses the Go modules to build and the only platform supported is Linux for production use. 
It will build and test on Mac OSX but only for development use.

### Q: Why do I get a "no such host" error resolving hostnames that other
programs can resolve?

Go uses a pure Go resolver by default for [name resolution](https://golang.org/pkg/net/#hdr-Name_Resolution).
This resolver behaves differently than the C library functions but is more
efficient when used with the Go runtime.

### Q: Can plugins from Telegraf be used

The intent of Telex is to not deviate too far from Telegraf so incorporating any plugins from Telegraf should be possible.
However, Telex has a goal to keep from growing the dependencies of subsequent libraries used by Telex and its plugins.
