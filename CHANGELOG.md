# Ferstream


<a name="v1.9.1"></a>
## [v1.9.1] - 2024-07-11
### Other Improvements
- fix vulnerablity libraries


<a name="v1.9.0"></a>
## [v1.9.0] - 2024-05-10
### Other Improvements
- upgrade nats client ([#23](https://github.com/kumparan/ferstream/issues/23))


<a name="v1.8.3"></a>
## [v1.8.3] - 2024-03-13
### New Features
- migrate to go.uber.org/mock


<a name="v1.8.2"></a>
## [v1.8.2] - 2022-12-14
### New Features
- update github action go.yml
- add github action for run unit test


<a name="v1.8.1"></a>
## [v1.8.1] - 2022-11-17

<a name="v1.8.0"></a>
## [v1.8.0] - 2022-11-02
### New Features
- new connection method with default config and reconnect handler


<a name="v1.7.0"></a>
## [v1.7.0] - 2022-10-28
### New Features
- subscribe ([#18](https://github.com/kumparan/ferstream/issues/18))


<a name="v1.6.1"></a>
## [v1.6.1] - 2022-10-11
### Fixes
- add log payload to help development ([#17](https://github.com/kumparan/ferstream/issues/17))


<a name="v1.6.0"></a>
## [v1.6.0] - 2022-10-07
### New Features
- add ParseNatsEventMessageFromBytes


<a name="v1.5.1"></a>
## [v1.5.1] - 2022-10-04
### Fixes
- revert parseFromBytes NatsEventMessage using fallbackWith

### Other Improvements
- add readme ([#12](https://github.com/kumparan/ferstream/issues/12))


<a name="v1.5.0"></a>
## [v1.5.0] - 2022-09-29

<a name="v1.4.0"></a>
## [v1.4.0] - 2022-09-28
### New Features
- add NatsEventAuditLogMessage ([#13](https://github.com/kumparan/ferstream/issues/13))


<a name="v1.3.5"></a>
## [v1.3.5] - 2022-08-31
### Fixes
- add nil checker for errHandler ([#11](https://github.com/kumparan/ferstream/issues/11))


<a name="v1.3.4"></a>
## [v1.3.4] - 2022-08-25
### Other Improvements
- upgrade go-utils


<a name="v1.3.3"></a>
## [v1.3.3] - 2022-08-24
### Fixes
- add missing method on MockJetStream


<a name="v1.3.2"></a>
## [v1.3.2] - 2022-08-23
### New Features
- add IDStr field

### Other Improvements
- fix error message


<a name="v1.3.1"></a>
## [v1.3.1] - 2022-07-27

<a name="v1.3.0"></a>
## [v1.3.0] - 2022-07-20
### New Features
- add consumer info and register method
- add message parser


<a name="v1.2.0"></a>
## [v1.2.0] - 2022-02-22
### New Features
- add tenantID in nats event message


<a name="v1.1.0"></a>
## [v1.1.0] - 2022-02-04
### New Features
- add subject in msg event


<a name="v1.0.2"></a>
## [v1.0.2] - 2022-01-24

<a name="v1.0.1"></a>
## [v1.0.1] - 2022-01-24

<a name="v1.0.0"></a>
## v1.0.0 - 2022-01-20
### New Features
- handle update config in AddStream method and update proto
- add method AddStream
- implement jetstream


[Unreleased]: https://github.com/kumparan/ferstream/compare/v1.9.1...HEAD
[v1.9.1]: https://github.com/kumparan/ferstream/compare/v1.9.0...v1.9.1
[v1.9.0]: https://github.com/kumparan/ferstream/compare/v1.8.3...v1.9.0
[v1.8.3]: https://github.com/kumparan/ferstream/compare/v1.8.2...v1.8.3
[v1.8.2]: https://github.com/kumparan/ferstream/compare/v1.8.1...v1.8.2
[v1.8.1]: https://github.com/kumparan/ferstream/compare/v1.8.0...v1.8.1
[v1.8.0]: https://github.com/kumparan/ferstream/compare/v1.7.0...v1.8.0
[v1.7.0]: https://github.com/kumparan/ferstream/compare/v1.6.1...v1.7.0
[v1.6.1]: https://github.com/kumparan/ferstream/compare/v1.6.0...v1.6.1
[v1.6.0]: https://github.com/kumparan/ferstream/compare/v1.5.1...v1.6.0
[v1.5.1]: https://github.com/kumparan/ferstream/compare/v1.5.0...v1.5.1
[v1.5.0]: https://github.com/kumparan/ferstream/compare/v1.4.0...v1.5.0
[v1.4.0]: https://github.com/kumparan/ferstream/compare/v1.3.5...v1.4.0
[v1.3.5]: https://github.com/kumparan/ferstream/compare/v1.3.4...v1.3.5
[v1.3.4]: https://github.com/kumparan/ferstream/compare/v1.3.3...v1.3.4
[v1.3.3]: https://github.com/kumparan/ferstream/compare/v1.3.2...v1.3.3
[v1.3.2]: https://github.com/kumparan/ferstream/compare/v1.3.1...v1.3.2
[v1.3.1]: https://github.com/kumparan/ferstream/compare/v1.3.0...v1.3.1
[v1.3.0]: https://github.com/kumparan/ferstream/compare/v1.2.0...v1.3.0
[v1.2.0]: https://github.com/kumparan/ferstream/compare/v1.1.0...v1.2.0
[v1.1.0]: https://github.com/kumparan/ferstream/compare/v1.0.2...v1.1.0
[v1.0.2]: https://github.com/kumparan/ferstream/compare/v1.0.1...v1.0.2
[v1.0.1]: https://github.com/kumparan/ferstream/compare/v1.0.0...v1.0.1
