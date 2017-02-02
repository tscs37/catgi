# v0.1.4:

## New Features:
* Generic GC that can be used by most backends without modification
* Implemented a basic compliancy test suite to run against backends
* Added Basis for expandable Backends in form of OnionBackend interface
* Generic Configuration Parser, used only in S3 backend atm
* A new LocalFS backend to store files in the filesystem (INNNOVATION!)
* A new AWS S3 backend that is untested (STABILITY!)
* Index.html and login.html are now embedded, when the folder `./resources` is present, that will be used instead of any embedded files. *Note* that this is a breaking change if you modified the index.html or login.html, however, it should be fairly quickly fixed
* When the IgnoreAuth option is set in the config, then logins are optional but logins still work
* Using http.ServeContent to properly serve files with mimedata and all, HTML files now werk!
* Generic Resource handler to serve any file from `./resource`

## Fixes & Notes:
* Proper package names for B2 Backend
* B2 Backend upload is now more robust
* Moved B2 naming functions into common package, these should be used by all packages in the future
* B2 ListGlob now actually works
* Backends now correct flake errors and don't panic on nil-Files anymore
* BuntDB's ListGlob should also work now
* Compliancy test suite for FCache, LocalFS and BuntDB
* Split up the common package over several Files
* Added tests to the common package (fixed some bugs there too)
* Simplified NewDriver in the backend package
* FCache bugs fixed that lead to crash on delete
* Password Validation is now in a dedicated package and allows for Dropbox like authentication
* HTTP Routes now use localized config instead of a global variable
* Logger can set output target from context
* Backends can now put arbitrary cookies and headers into response
* Additionally backends can also take over writing the HTTP response
* Started implementing a crypto backend for future use
* Simplified ID Fountain Code
* Context will now contain the HTTP request
* Password verification has been moved into a dedicated file
* Updated vendor.json

## Non-Code-Related:
* Using Drone + Travis now, two CI's are better than one, right?
* Improved Travis CI pipeline
* Added Backend List to README
* You'll find some development notes in /papers now, atm this concerns mainly cryptography stuff
# v0.1.3:

## New Features:
* The Server now catches Signals from the OS and handles them (somewhat) properly. Please see the code for licensing details.

## Fixes & Notes:
* Moved the backend/types code to backend/common, which will be expanded on
* FCache will now log on .Exists for clarity when it errors
* Basis for LocalFS backend is put in but not active yet, this should not affect the builds
* BuntDB now defaults to in-memory and AutoTTL enabled when no params specified
* The Logger will insert a request ID into the context for better logging context
* The Request ID is a snowflake style ID which is generated based on timestamp.
* Removed some unneeded dependencies from the vendor.json 

## Non-Code-Related:
* Travis now deploys directly to Github
* README links to travis correctly
* The Redirect Functionality was merged into develop, mistake on my part.

# v0.1.2:

## New Features:
* Added Garbage Collector HTTP Endpoint
* B2 Backend has a Garbage Collector now
* BuntDB Backend has a Garbage Collector now
* BuntDB can now disable internal TTL through config options
* BuntDB shrinks the DB after a GC-Run, if the DB is not in memory, it will report the number of bytes saved on the log (debug level)
* Makepass now supports DropBox passwords, however this is not supported yet on Catgi (Makepass will be phased out in time)
* Catgi now recognized different authentication methods using a seperate string. It will fall back to legacy authentication if no type is specified.


## Fixes & Notes:
* Removed Publisher Code from B2 Backend
* Publishing has been fully removed so it can be added later in a more effective manner
* Small Fix to ListGlob, preventing it from actually working
* Added functions to check if a file is clpubName or pubName
* BuntDB performs Rollbacks on Error
* Code Error prevented BuntDB from working correctly
* BuntDB register logs errors for better debugging
* FCache will no longer panic after a GC or Removal of Files
* Removed unused interface types in drivers/types/types.go
* DateOnlyTime now uses a TTL() function to reliably determine a Time-To-Live

## Non-Code-Related:
* CatGiFS was removed and will be worked on seperately when I have time
* Removed Makepass Binary from Gitrepo, requires building now
* Added .travis.yml, builds are now tested and tested on push
* Added some stuff to the README

# v0.1.1:

## Features:
* ListGlob now actually works (but still unused) in all backends
* Split up Backend interfaces into Backend, Publisher, KVBackend and ContentBackend. Only Backend is actively used.
* File now contain meta about the owner in the User field
* Authentication is lazy on index page, attaching a header of the currently logged in user to the response
* Adding ?raw=1 to a file request will print the raw JSON of the file (meta + data)
* The File Handler can now handle /f/flake/name.ext, /f/flake/ and /f/flake in addition to /file/flake
* Catgi can now proxy to a Piwik instance for traffic analysis. This is somewhat buggy but it allows to record visits atleast

## Fixes & Notes:
* Added Github Mirror and Original Gitrepo links to README
* B2 checks if a file is meta or data in ListGlob
* FCache will no longer proxy calls to publisher interfaces
* Cleaned up in the Auth Checker
* Amended Message in InjectLog Handler
* Uploading Files now sets the "CreatedAt" / "created_at" fields
* Uploading Files no redirects to /f/flake/filename.extension
* Started some work to someday maybe allow FUSE mounts of Catgi Backends (CatgiFS for short)

# v0.1.0:

Initial Release