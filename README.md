# codex

Codex provides a historical context about devices connected to [XMiDT](https://github.com/Comcast/xmidt).

[![Build Status](https://travis-ci.com/Comcast/codex.svg?branch=master)](https://travis-ci.com/Comcast/codex)
[![codecov.io](http://codecov.io/github/Comcast/codex/coverage.svg?branch=master)](http://codecov.io/github/Comcast/codex?branch=master)
[![Code Climate](https://codeclimate.com/github/Comcast/codex/badges/gpa.svg)](https://codeclimate.com/github/Comcast/codex)
[![Issue Count](https://codeclimate.com/github/Comcast/codex/badges/issue_count.svg)](https://codeclimate.com/github/Comcast/codex)
[![Go Report Card](https://goreportcard.com/badge/github.com/Comcast/codex)](https://goreportcard.com/report/github.com/Comcast/codex)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/Comcast/codex/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/Comcast/codex.svg)](CHANGELOG.md)

## Summary

Codex accepts incoming events, stores them in a postgres database, and 
provides event information by device id.  This repo is a library of packages 
used to implement codex.

## The Pieces

<img src="./docs/images/flow.png" width=720 />

* **Database:** Any postgres database will work.  In `deploy/`, cockroachdb is 
  used.
* **[Svalinn](https://github.com/Comcast/codex-svalinn):** Registers to an 
  endpoint to receive events (Optional).  Has an endpoint that receives events
  as [WRP Messages](https://github.com/Comcast/wrp-c/wiki/Web-Routing-Protocol),
  parses them, and inserts them into the database.
* **[Gungnir](https://github.com/Comcast/codex-gungnir):** Has endpoints that 
  provide device information from the database.
* **[Fenrir](https://github.com/Comcast/codex-fenrir):** Deletes old records 
  from the database at an interval.

## Install
This repo is a library of packages used for the codex project.  There is no 
installation.  To install each service, go to their respective READMEs.

## Deploy
for deploying the project in Docker, refer to the deploy [README](deploy/README.md).

## Contributing
Refer to [CONTRIBUTING.md](CONTRIBUTING.md).