# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.2.7]
 - Removed creating the events table, now we just verify that it exists
 - Added limit parameter for finding records from the database

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

[Unreleased]: https://github.com/Comcast/codex/compare/v0.2.7...HEAD
[v0.2.7]: https://github.com/Comcast/codex/compare/v0.2.6...v0.2.7
[v0.2.6]: https://github.com/Comcast/codex/compare/v0.2.5...v0.2.6
[v0.2.5]: https://github.com/Comcast/codex/compare/v0.2.4...v0.2.5
[v0.2.4]: https://github.com/Comcast/codex/compare/v0.2.3...v0.2.4
[v0.2.3]: https://github.com/Comcast/codex/compare/v0.2.2...v0.2.3
[v0.2.2]: https://github.com/Comcast/codex/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/Comcast/codex/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/Comcast/codex/compare/v0.1.5...v0.2.0
[v0.1.5]: https://github.com/Comcast/codex/compare/v0.1.4...v0.1.5
[v0.1.4]: https://github.com/Comcast/codex/compare/v0.1.3...v0.1.4
[v0.1.3]: https://github.com/Comcast/codex/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/Comcast/codex/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/Comcast/codex/compare/v0.1.0...v0.1.1
[v0.1.0]: https://github.com/Comcast/codex/compare/0.0.0...v0.1.0
