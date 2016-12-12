# Fast Geocoder

Small script to geocode IP's using maxmind free database

## Installation

_requires go_

Clone the repo and build with
 [go build](https://golang.org/pkg/go/build/)

Or just download the compiled binary for mac or linux in the bin folder (no windows yet, so _contribute_)

## Usage

You may use this tool in 3 different ways (so far)

### Via command line
Using the command line, you can execute de geocode of a single IP address or passing a CSV file, geocode them all!

####Using the command line####

After building the project (or downloading the binaries), you can execute the following
 
 fast_geocoder _IP_ADDRESS_

or

 fast_geocoder _path-to-some-csv-file_

Both will print the result for the IP (or IPs) address you provided

The CSV file must contain a IP address string on the first column (only constraint)

### Via http server

You can also run a http server that can return the same results by calling the service on a provided port
 
 fast_geocoder --server -port 80

Port is optional, it will run on port 8080 by default. Have in mind you must have the permissions to run on the provided port or the server wont run.

On the example above, you can make calls to
 
 http://localhost/geocode/___IP_ADDRESS___

and receive a response on the format

 _{"lat": -33.4625, "lng": -70.6682, "country": "CL", "timezone": "America/Santiago"}_

## Contribution

Fork, code, push, PR :)

## License

> Copyright © 2016 [Thiago Costa](mailto:thiago@arroadie.com)
> This work is free. You can redistribute it and/or modify it under the
> terms of the Do What The Fuck You Want To Public License, Version 2,
> as published by Sam Hocevar. See the COPYING file for more details.