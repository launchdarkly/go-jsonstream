# Change log

All notable changes to the project will be documented in this file. This project adheres to [Semantic Versioning](http://semver.org).

## [3.1.3](https://github.com/launchdarkly/go-jsonstream/compare/v3.1.2...v3.1.3) (2026-08-24)


### Bug Fixes

* Bump minimum Go version to 1.26 ([#63](https://github.com/launchdarkly/go-jsonstream/issues/63)) ([5416a5c](https://github.com/launchdarkly/go-jsonstream/commit/5416a5cf866509e878ad281f9a770c60740353a4))

## [3.1.2](https://github.com/launchdarkly/go-jsonstream/compare/v3.1.1...v3.1.2) (2026-08-13)


### Bug Fixes

* Improve jreader throughput with a single-pass tokenizer ([#58](https://github.com/launchdarkly/go-jsonstream/issues/58)) ([78d0ac9](https://github.com/launchdarkly/go-jsonstream/commit/78d0ac92ec0f3a52f5f353529da13aeca93138d4))
* Improve jwriter throughput with append-based buffer internals ([#56](https://github.com/launchdarkly/go-jsonstream/issues/56)) ([acb84f4](https://github.com/launchdarkly/go-jsonstream/commit/acb84f48fcfba1eee2f8754d6d281764c6826a52))

## [3.1.1](https://github.com/launchdarkly/go-jsonstream/compare/v3.1.0...v3.1.1) (2026-02-20)


### Bug Fixes

* Bump gopkg.in/yaml.v3 from 3.0.0 to 3.0.1 ([#30](https://github.com/launchdarkly/go-jsonstream/issues/30)) ([261baa5](https://github.com/launchdarkly/go-jsonstream/commit/261baa546767497d371c4a20d873e685728b9ab6))
* Bump minimum go to 1.24 ([#34](https://github.com/launchdarkly/go-jsonstream/issues/34)) ([4538575](https://github.com/launchdarkly/go-jsonstream/commit/45385758baecad4ec2dbc4e6b42a47ba1e92ff7f))

## [3.1.0] - 2024-01-18
### Added:
- Adds a new `StringAsBytes()` method, which can be used instead of the always-allocating `String()` method.


### Changed:
- GC improvement: in non-easyjson builds, when SkipValue encounters strings, allocation is eliminated. Thanks, @bobby-stripe!

## [3.0.0] - 2022-08-29
This release drops compatibility with Go 1.17 and below, and changes the import path from `github.com/launchdarkly/go-jsonstream/v2` to `github.com/launchdarkly/go-jsonstream/v3`. There are no other changes.

## [2.0.0] - 2022-03-18
This release drops compatibility with Go 1.15 and below, and changes the import path from `gopkg.in/launchdarkly/go-jsonstream.v1` to `github.com/launchdarkly/go-jsonstream/v2`. There are no functional changes.

## [1.0.1] - 2021-06-03
### Fixed:
- Parsing of numeric values in the default implementation was broken for numbers that have an exponent but do not have a decimal (such as 1e-5, as opposed to 1.0e-5). For such numbers, the parser was returning an integer value based on misusing the ASCII values of the non-digit characters as if they were digits, e.g. 1e-5 was interpreted as 88035. This bug did not occur in the EasyJSON implementation of the parser.

## [1.0.0] - 2020-12-17
Initial release of this library.
