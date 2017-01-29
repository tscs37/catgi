# Erasure Encoding in Catgi

Using Erasure Encoding with Checksumming, Catgi can ensure the following
properties for a replicated storage backend:

* Fault-Tolerant for m-of-k
* Bitrot-Tolerant for m-of-k
* Byzantine-Tolerant for m-of-k using HMAC
* Backward Compatible if new backends are added

## Configuration

* `mode` Encoding Mode
* `allowed_fails` Number of Tolerated Storage Failures
* `confirm_factor` Confirmation Factors
* optional `index` Index Backend
* optional `retrieve_all`

The Backend will determine the number of data and checksum shards
automatically based on the number of backends. The number of data shards
is set as `backends - allowed_fails`

The Confirmation Factor defines how many backends over the failure-tolerance
must be written to. This is determined as `data_shards + confirm_factor`

The Index Backend is used for metadata. If not specified metadata is instead
replicated across all backends.

`retrieve_all` is a boolean value that indicates how the backend proceeds after
reassembling a file.

### Examples

#### 3 Backends, 1 Allowed Fail, 0 Confirm Factor

* 2 Data Shards + 1 Erasure Shard
* 2 Writes for Success

#### 30 Backends, 8 Allowed Fail, 3 Confirm Factor

* 22 Data Shards + 8 Erasure Shards
* 25 Writes for Success

## Encoding

Encoding happens in two modes;

* HCCA - Hash-Chunk Content Addressed
* SCFA - Single-Chunk File Addressed

The first mode splits a file into several chunks of a maximum size and
uses erasure encoding on each part seperately. This allows for deduplication
and content-addressing across several backends.

The second mode uses erasure encoding on a whole file at once.

### HCCA

By default a 4KB chunk size is chosen.

The files are chunked according to this maximum.

Each chunk is erasure encoded and then distributed across the backends.

The metadata file contains an ordered hash list of the shards and is stored
either over all backends or in the index backend.

If a chunk is smaller than the rest, then it's padded with zeroes.

### SCFA

The file as a whole is erasure encoded and then spread over the backends.

The metafile is replicated to all backends or uses the index backend.

### Encoding

After the first mode step, each mode operates equally.

The incoming data is sharded according to the configuration.

The data is encoded using msgpacks, the field `data` contains the
shard data.

The field `hash` contains the Blake2b hash of the `data` field.

The field `hmac` contains a HMAC using Blake2b and a serverside key.

The field `data_shards` contains the number of data shards used.

The field `parity_shards` contains the number of parity shards used.

### Reading Files

After retrieving the metadata file the backend attempts to read
the file from each backend in turn (or several at once).

The backend will determine the parity configuration from these files
as it may have changed in the past.

If it has changed, the erasure encoding will be updated and the file stored
into the backend again.

If backends have been removed then the backend will abort if the number of
backends is lower than the number of data shards required.

The hash and hmac of the file will be calculated and compared, if it mismatches
the shard is treated as missing and will be deleted.

The moment enough shards have been collected, the file is reassembled
and then returned to the frontend.

If `retrieve_all` is set then the remaining shards will be retrieved in a 
background routine and validated.

And missing shards (read: corrupt and missing shards) will be rewritten to the
backend they belong to. This is done by reassembling the file and recreating
it.
