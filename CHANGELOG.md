# Change log

All notable changes to the project will be documented in this file. This project adheres to [Semantic Versioning](http://semver.org).

## [1.0.1] - 2021-06-03
### Fixed:
- Parsing of numeric values in the default implementation was broken for numbers that have an exponent but do not have a decimal (such as 1e-5, as opposed to 1.0e-5). For such numbers, the parser was returning an integer value based on misusing the ASCII values of the non-digit characters as if they were digits, e.g. 1e-5 was interpreted as 88035. This bug did not occur in the EasyJSON implementation of the parser.

## [1.0.0] - 2020-12-17
Initial release of this library.
