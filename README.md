# props: Go (golang) library for handling Java-style property files and app configuration

This library provides compatibility with Java property files for Go.

It also provides additional features such as property expansion, configuration 
by convention, and encryption for use as application configuration. The approach
is similar to that of the Environment classes of Spring Boot.

## Important Types
There are three main types provided:
* `Properties` - read and write property files in Java format
* `Expander` - replaces property references wrapped by '${}' 
(or other custom prefix/suffix) at runtime (as found in Ant/Log4J/JSP EL/Spring)
* `Configuration` - provides easy configuration by convention, parsing of
property values into Go types, and encryption support

The full Java property file format including all comment types, line 
continuations, key-value separators, unicode escapes, etc. is supported.

## Configuration by Convention
The standard convention supported provides for profile and environment specific
property files along with command line arguments and environment variables.

Properties are resolved in the following priority order:
1. Command line arguments
1. Environment variables
1. `<prefix>-<profile>.properties` for the provided prefix and profiles values 
(in order)
1. `<prefix>.properties` for the provided prefix value

The first matching property value found will be returned.

## Custom Configuration
The types provided can be included or excluded in any order to create an 
alternative configuration.
* `Arguments`
* `Environment`
* `Expander`
* `Properties`

Combine multiple property source lookups with the `Combined` type.

## Command Line Utility
A command line utility is provided in the `cmd` directory. This app is used to
encrypt, decrypt, or re-encrypt property files or individual values.

## Encryption
Encryption is handled by putting a marker prefix (`[enc:x]`) on encrypted 
values. The prefix indicates which algorithm was used for encryption and allows 
for different algorithms to co-exist or be upgraded at different times in the 
same file.

The standard approach is to update your property file with plaintext values
with the `[enc:0]` prefix, for example:

`db.password=[enc:0]$ecr3t!`

Then run the `encryptFile` command from app in the `cmd` dir to convert the
result into an encrypted value:

`db.password=[enc:1]<base64 data>`
