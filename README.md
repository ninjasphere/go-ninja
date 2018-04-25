# Ninja Sphere - Golang Library

[![godoc](http://img.shields.io/badge/godoc-Reference-blue.svg)](https://godoc.org/github.com/ninjasphere/go-ninja)
[![MIT License](https://img.shields.io/badge/license-MIT-yellow.svg)](LICENSE)
[![Ninja Sphere](https://img.shields.io/badge/built%20by-ninja%20blocks-lightgrey.svg)](http://ninjablocks.com)
[![Ninja Sphere](https://img.shields.io/badge/works%20with-ninja%20sphere-8f72e3.svg)](http://ninjablocks.com)

[![forthebadge](https://forthebadge.com/images/badges/built-by-hipsters.svg)](http://forthebadge.com)
---


### Introduction
A Golang library to interact with the Ninja Sphere-- used for creating tubular new drivers and radical apps.

It takes care of the connection to MQTT, implementing and calling services using JSON-RPC, device protocols, schema validation, configuration, logging etc. It's your hero in a half-sphere.

*tl;dr - Writing a driver in Go? Use this.*

![go ninja go ninja go](http://cdn3.whatculture.com/wp-content/uploads/2013/05/vanilla-ice-ninja-turtles.jpg)


For development outside of a devkit/sphere, ensure you have sphere-serial in your path, and have sphere-config and schemas checked out and accessible.

### Usage

For example usage in drivers, check out [driver-go-chromecast](https://github.com/ninjasphere/driver-go-chromecast), the example [fakedriver](fakedriver) or any of our other released drivers.

### See Also
- [schemas](https://github.com/ninjasphere/schemas) - Json Schemas describing all the communication between services (drivers, apps etc.) in Ninja Sphere.

- [sphere-config](https://github.com/ninjasphere/sphere-config) - The base configuration that is shared in Sphere.

### More Information

More information can be found on the [project site](https://github.com/ninjasphere/go-ninja) or by visiting the Ninja Blocks [forums](https://discuss.ninjablocks.com).

### Contributing Changes

To contribute code changes to the project, please clone the repository and submit a pull-request ([What does that mean?](https://help.github.com/articles/using-pull-requests/)).

### License
This project is licensed under the MIT license, a copy of which can be found in the [LICENSE](LICENSE) file.

### Copyright
This work is Copyright (c) 2014-2015 - Ninja Blocks Inc.
