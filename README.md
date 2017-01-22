# CatGi - Selfhosted Image and Code Dump

[![Travis](https://img.shields.io/travis/tscs37/catgi.svg?style=flat-square)](https://travis-ci.org/tscs37/catgi)


CatGi is a pluggable, extensible and selfhosted provider of images, text and more.

The App does not rely on any configuration and operates
as an in-memory testable filestore until a config is specified.

At the moment files can be stored for up to 1 month and no longer.
Permanent storage is a goal for a future release.

## Usage

CatGi requires [govendor](https://github.com/kardianos/govendor) to be build.

```
# Clone the repository
git clone gogs@git.timschuster.info:rls.moe/catgi 
# Go into Repository
cd catgi

# Optional:
# Install missing dependencies
govendor fetch +missing

# Optional:
# Install full vendor directory
govendor fetch +all

# Otherwise
# Install missing dependencies globally
go get

# Build makepass
cd makepass
go build

# Build catgi
cd ..
cd catgi
go build
```

### makepass

`makepass` is a simple utility to generate user configuration.

It asks for user and password (no confirmation) on the CLI and
then prints a JSON string that can be copy-pasted into the configuration

### catgi

```
catgi [config-file]
```

Starts the webserver and configured backends. If no config is specified,
catgi starts in default configuration.

The default configuration stores everything in memory and leaves no
permanent traces on the system (usually) and requires no login.

It is recommended to setup authentication.

## Configuration

Here is an example config file:

```
{
    "backend": {
        "driver": "fcache",
        "params": {
            "driver": "buntdb",
            "params": {
                "file": ":memory:"
            },
            "cache_size": 20,
            "async_upload": false
        }
    },
    "users": [
    ],
    "http": {
        "port": 8080,
        "listen": "[::1]"
    },
    "loglevel": "debug"
}
```

This configuration will setup CatGi as follows:

* Set highest logging level : `loglevel`
* Listen on `[::1]:8080` for HTTP traffic : `http.port` and `http.listen`
* Empty user list means no login required : `users`
* Use fcache for the backend : `backend.driver`
    * fcache will use buntdb as backend : `backend.params.driver`
    * fcache will cache 20 entries : `backend.params.cache_size`
    * fcache will wait for upload to complete : `backend.params.async_upload`
        * buntdb will use an in-memory db

Other backends might required differing configuration.

### Available backends

| Name         | Driver Name | Notes                                |
|--------------|-------------|--------------------------------------|
| B2Backblaze  | `b2`        | No automatic GC and rather slow      |
| BuntDB       | `buntdb`    | Automatic GC and fast                |
| FCache       | `fcache`    | Caching Backend, not standalone      |
| LocalFS      | `localfs`   | Stores in Filesystem, no auto GC     |

## License

CatGi is licensed under MPL 2.0

Dependencies are under their respective license and copyright.

## Contribute

Pull requests should be well formatted.

HTML should be kept to minimal filesize, CSS or JS should be avoided.

Pull Request will be accepted from any of the code mirrors.

## Code Repository

[Origin](https://git.timschuster.info/rls.moe/catgi)

[Github](https://github.com/tscs37/catgi)

## Future

I regard CatGi as mostly feature complete for myself missing only two
things:

    * Public Gallery
    * Automatic Garbage Collection (Manual GC works now)
    * Permanent Files
    * Named Publishing (eg `catgi.rls.moe/n/helloworld.txt`)

I have a few other features planned to, like S3 and GCS support,
federating across servers and much more.

Pull Requests for additional features are welcome.