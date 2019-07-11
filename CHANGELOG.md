# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.11.0]
- Separated out library packages, so only deploy and tests are left.

## [v0.10.0]
- Added `capacityset` package
- Changed `batchDeleter` to leverage `capacityset`

## [v0.9.0]
- Changed batchDeleter to expect the deathdate to be in Unix Nanoseconds

## [v0.8.0]
- Refactored pruning
- Added comments for godocs
- Added a way to get a list of device ids

## [v0.7.0]
- Added `record_id` to index
- Changed `minMaxBatchSize` from 0 to 1
- Restructured `db` repo

## [v0.6.0]
- Added documentation
- Added prometheus to docker-compose
- Fixed batching on insert
- Added `BatchDeleter`, which gets a list of record ids that have passed their 
  deathdate and deletes them in batches.

## [v0.5.0]
- Removed primary key id record

## [v0.4.3]
- Added some metrics

## [v0.4.2]
- Changed test structure
- Added defaults to replace invalid config values
- Removed list of events from insert error

## [v0.4.1]
- fixed cipher yaml loading

## [v0.4.0]
- Added kid and alg to `db` package
- Updated loading cipher options

## [v0.3.3]
- Fix box loader in `cipher` package

## [v0.3.2]
- Revamped `cipher` package
- Added box algorithm to `cipher` package
- Added regex support to `blacklist`
- Made EventType a stringer

## [v0.3.1]
- Added `blacklist` package

## [v0.3.0]
- Added Fenrir to docker-compose and related files
- Added basic cipher file loading, with viper
- Added prune limit
- Removed `Event` struct from `db` package

## [v0.2.9]
 - Added logger wrapper for health

## [v0.2.8]
 - Fixed limit parameter

## [v0.2.7]
 - Removed creating the events table, now we just verify that it exists
 - Added limit parameter for finding records from the database
 - replaced dep with modules

## [v0.2.6]
 - Fixed birthdate and deathdate in record schema

## [v0.2.5]
 - Modified record schema in `db` package

## [v0.2.4]
 - Modified record schema in `db` package

## [v0.2.3]
 - Fix metrics

## [v0.2.2]
 - Added SQL query success, failure, and retry metrics
 - Added metric for number of rows deleted

## [v0.2.1]
 - Toned down travis
 - Updated comments for swagger docs
 - Removed the index on type
 - Added multi record insert support

## [v0.2.0]
 - Added `cipher` package
 - Fixed `db` package mocks and unit tests to match cockroachdb code
 - Fixed `db` package error statements to be more accurate
 - Adding metrics to `db` package
 - Added _ping_ and _close_ to `db` package
 - Modified db timeouts to take time.Duration values
 - Updated swagger comments to include examples

## [v0.1.5]
 - Added event type to record in `db` package

## [v0.1.4]
- Changed `db` package to use cockroachdb instead of couchbase

## [v0.1.3]
- Added retry decorators for inserting, pruning, and getting

## [v0.1.2]
- Allowed parametization of tags for docker
- Simplified travis.yaml file
- Restructured `db` package for cleaner unit tests
- Added unit tests for `db` and `xvault` packages

## [v0.1.1]
- Modified `db` package database schema to have Tombstone and History documents
- Added files for docker-compose
- Added initial cucumber tests
- Added initial unit tests for `db` package

## [v0.1.0]
- Initial creation
- Created `db` and `xvault` package

[Unreleased]: https://github.com/xmidt-org/codex-deploy/compare/v0.11.0..HEAD
[v0.11.0]: https://github.com/xmidt-org/codex-deploy/compare/v0.10.0...v0.11.0
[v0.10.0]: https://github.com/xmidt-org/codex-deploy/compare/v0.9.0...v0.10.0
[v0.9.0]: https://github.com/xmidt-org/codex-deploy/compare/v0.8.0...v0.9.0
[v0.8.0]: https://github.com/xmidt-org/codex-deploy/compare/v0.7.0...v0.8.0
[v0.7.0]: https://github.com/xmidt-org/codex-deploy/compare/v0.6.0...v0.7.0
[v0.6.0]: https://github.com/xmidt-org/codex-deploy/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/xmidt-org/codex-deploy/compare/v0.4.3...v0.5.0
[v0.4.3]: https://github.com/xmidt-org/codex-deploy/compare/v0.4.2...v0.4.3
[v0.4.2]: https://github.com/xmidt-org/codex-deploy/compare/v0.4.1...v0.4.2
[v0.4.1]: https://github.com/xmidt-org/codex-deploy/compare/v0.4.0...v0.4.1
[v0.4.0]: https://github.com/xmidt-org/codex-deploy/compare/v0.3.3...v0.4.0
[v0.3.3]: https://github.com/xmidt-org/codex-deploy/compare/v0.3.2...v0.3.3
[v0.3.2]: https://github.com/xmidt-org/codex-deploy/compare/v0.3.1...v0.3.2
[v0.3.1]: https://github.com/xmidt-org/codex-deploy/compare/v0.3.0...v0.3.1
[v0.3.0]: https://github.com/xmidt-org/codex-deploy/compare/v0.2.9...v0.3.0
[v0.2.9]: https://github.com/xmidt-org/codex-deploy/compare/v0.2.8...v0.2.9
[v0.2.8]: https://github.com/xmidt-org/codex-deploy/compare/v0.2.7...v0.2.8
[v0.2.7]: https://github.com/xmidt-org/codex-deploy/compare/v0.2.6...v0.2.7
[v0.2.6]: https://github.com/xmidt-org/codex-deploy/compare/v0.2.5...v0.2.6
[v0.2.5]: https://github.com/xmidt-org/codex-deploy/compare/v0.2.4...v0.2.5
[v0.2.4]: https://github.com/xmidt-org/codex-deploy/compare/v0.2.3...v0.2.4
[v0.2.3]: https://github.com/xmidt-org/codex-deploy/compare/v0.2.2...v0.2.3
[v0.2.2]: https://github.com/xmidt-org/codex-deploy/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/xmidt-org/codex-deploy/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/xmidt-org/codex-deploy/compare/v0.1.5...v0.2.0
[v0.1.5]: https://github.com/xmidt-org/codex-deploy/compare/v0.1.4...v0.1.5
[v0.1.4]: https://github.com/xmidt-org/codex-deploy/compare/v0.1.3...v0.1.4
[v0.1.3]: https://github.com/xmidt-org/codex-deploy/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/xmidt-org/codex-deploy/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/xmidt-org/codex-deploy/compare/v0.1.0...v0.1.1
[v0.1.0]: https://github.com/xmidt-org/codex-deploy/compare/0.0.0...v0.1.0
