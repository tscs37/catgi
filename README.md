# CatGi - Selfhosted Image and Code Dump

CatGi is a pluggable, extensible and selfhosted provider of images, text and more.

## Storage

CatGi's backend is kept as simple as possible so a wide
range of platforms can be supported.

Therefore the only methods the backend implements are the following:

* Upload
* Exists
* Get
* Delete
* LoadIndex
* StoreIndex
* ListGlob

The requirements for the underlying storage are as follows:

* Create new files and store contents with certain names
* Check if a file of a certain name exists 
* Retrieve a file based on a name
* Delete a file based on a name
* List all files matching a certain name glob (ie "file/*")

The LoadIndex and StoreIndex functions *can* be wrappers
around Upload/Exists/Get/delete or they can also just call
the storage directly. Either way works.

The Two Index functions must merely guarantee that the most
recent version of the index can be retrieved.

## Index

The Index is the heart of CatGi storage. It keeps track
of the lifetime of file objects, their metadata and
caches data.

It's function set almost resembles HTTP Endpoints (Get, Put and Delete)
but not quite yet.

The Index interfaces with a given backend directly, so one index
can theoretically be used for several backends at the same time
as long as these backends are consistent or the index allows
such operation.

The purpose of the index is to keep a cached version of the
backend around.

