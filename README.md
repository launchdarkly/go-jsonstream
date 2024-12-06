# LaunchDarkly Streaming JSON for Go

[![Actions Status](https://github.com/launchdarkly/go-jsonstream/actions/workflows/ci.yml/badge.svg?branch=v3)](https://github.com/launchdarkly/go-jsonstream/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/launchdarkly/go-jsonstream/v3.svg)](https://pkg.go.dev/github.com/launchdarkly/go-jsonstream/v3)

## Overview

The `go-jsonstream` library implements a streaming approach to JSON encoding and decoding which is more efficient than the standard mechanism in `encoding/json`. Unlike `encoding/json` or other reflection-based frameworks, it has no knowledge of structs or other complex types; you must explicitly tell it what values and properties to write or read. It was implemented for the [LaunchDarkly Go SDK](https://github.com/launchdarkly/go-server-sdk) and other LaunchDarkly Go components, but may be useful in other applications.

There are two possible implementations, selectable via build tags:

1. A default implementation that has no external dependencies, compatible with all platforms. This performs better than `encoding/json`, but not as well as the other two below.

2. An implementation that uses the low-level tokenizing and output functions from the [easyjson](https://github.com/mailru/easyjson) library (but without the code generation mechanism that easyjson also provides). This is used if you enable the build tag `launchdarkly_easyjson`.

Although the easyjson implementation is the fastest, it is opt-in rather than being the default, for two reasons:

* By default, easyjson uses Go's `unsafe` package, which may be undesirable or not allowed depending on your runtime environment. Easyjson disables the use of `unsafe` if you set the build tag `easyjson_nounsafe` or `appengine`.

* Although easyjson is widely used, at this time it does not have a stable 1.x version.

The design of `go-jsonstream` allows encoding/decoding logic to be written against the same API without needing to know whether the easyjson implementation will be used at build time or not. There is an adapter (also conditionally compiled with the `launchdarkly_easyjson` tag) to allow `go-jsonstream` to plug directly into a `jlexer.Lexer` or `jwriter.Writer` that is being used to read or write some other type with easyjson.

The unit tests for `go-jsonstream` define a common test suite that is run against the default implementation and the easyjson implementation, to verify that their behavior is consistent across a large number of permutations of possible JSON inputs and outputs.

## Supported Go versions

This version of the project requires a Go version of 1.18 or higher.

## Contributing

We encourage pull requests and other contributions from the community. Check out our [contributing guidelines](CONTRIBUTING.md) for instructions on how to contribute to this SDK.

## About LaunchDarkly

* LaunchDarkly is a continuous delivery platform that provides feature flags as a service and allows developers to iterate quickly and safely. We allow you to easily flag your features and manage them from the LaunchDarkly dashboard.  With LaunchDarkly, you can:
    * Roll out a new feature to a subset of your users (like a group of users who opt-in to a beta tester group), gathering feedback and bug reports from real-world use cases.
    * Gradually roll out a feature to an increasing percentage of users, and track the effect that the feature has on key metrics (for instance, how likely is a user to complete a purchase if they have feature A versus feature B?).
    * Turn off a feature that you realize is causing performance problems in production, without needing to re-deploy, or even restart the application with a changed configuration file.
    * Grant access to certain features based on user attributes, like payment plan (eg: users on the ‘gold’ plan get access to more features than users in the ‘silver’ plan). Disable parts of your application to facilitate maintenance, without taking everything offline.
* LaunchDarkly provides feature flag SDKs for a wide variety of languages and technologies. Check out [our documentation](https://docs.launchdarkly.com/docs) for a complete list.
* Explore LaunchDarkly
    * [launchdarkly.com](https://www.launchdarkly.com/ "LaunchDarkly Main Website") for more information
    * [docs.launchdarkly.com](https://docs.launchdarkly.com/  "LaunchDarkly Documentation") for our documentation and SDK reference guides
    * [apidocs.launchdarkly.com](https://apidocs.launchdarkly.com/  "LaunchDarkly API Documentation") for our API documentation
    * [blog.launchdarkly.com](https://blog.launchdarkly.com/  "LaunchDarkly Blog Documentation") for the latest product updates
    * [Feature Flagging Guide](https://github.com/launchdarkly/featureflags/  "Feature Flagging Guide") for best practices and strategies
