# Change log

All notable changes to the project will be documented in this file. This project adheres to [Semantic Versioning](http://semver.org).

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
