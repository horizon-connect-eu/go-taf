# CONNECT Trust Assessment Framework

| Artifact Overview           |                                                           |
|-----------------------------|-----------------------------------------------------------|
| Released Artifact           | Trust Assessment Framework                                |
| Identifier                  | `github.com/horizon-connect-eu/go-taf`                    |
| License                     | Apache-2.0 license                                        |
| Work Package                | 3                                                         |
| Responsible Project Partner | Institute of Distributed Systems, Ulm University, GERMANY |


This repository provides the final prototype of the Go-based Trust Assessment Framework released as part of the Horizon CONNECT project.

## Build from Source

First, clone this repository:
```shell
git clone git@github.com:horizon-connect-eu/go-taf.git
```
Next, go to the `go-taf` directory and run make:

```shell
cd go-taf
make build
```

To run the TAF, you can also use make: 

```shell
make run
```

To build and run the TAF with an enabled debugging webinterface, you can use the following make command:

```shell
GOFLAGS=-tags=webui make run
```


## Configuration

The TAF uses an internal configuration with hardcoded defaults. To change the configuration, you can use a JSON file (template located in `res/taf.json`) and specify the actual file location in the environment variable `TAF_CONFIG`. The following options can be configured. Missing options are implicitly using their defined default values.

```js
{
  "Identifier": "taf",                  // internal identifier of this instance 
  "Communication": {
    "Kafka": {
      "Broker": "localhost:9092",       // address and port of the kafka bootstrap server
      "TafTopic": "taf"                 // kafka topic the TAF will consume
    },
    "TafEndpoint": "taf",               // kafka identifier of TAF component
    "AivEndpoint": "aiv",               // kafka identifier of AIV component
    "MbdEndpoint": "mbd"                // kafka identifier of MBD component
  },
  "Logging": {
    "LogLevel": 2,                      // log level: 1=TRACE, 2=DEBUG, 3=INFO,
                                        //    4=WARN, 5=ERROR, 6=FATAL, 7=PRINT
    "LogStyle": "PRETTY"                // log style: 'PRETTY', 'JSON', or 'PLAIN'
  },
  "Crypto": {
    "Enabled": true,                    // whether the crypto library should be used or not
    "KeyFolder": "res/cert/",           // path to key folder that is passed to crypto library
    "IgnoreVerificationResults": false  // false: discard messages that failed to verify
                                        // true: process messages that failed to verify
                                        //        (a warning will be logged to console)
  },
  "Debug": {
    "FixedSessionID": "",               // if provided, this fixed value is used by the TAM
                                        // instead of a random UUID-based session id
    "FixedSubscriptionID": "",          // if provided, this fixed value is used by the TAM
                                        // instead of a random UUID-based subscription id
    "FixedRequestID": ""                // if provided, this fixed request id is used by the
                                        // trust source manager instead of a random UUID-based id
  },
  "Evidence": {
    "AIV": {
      "CheckInterval": 1000             // check interval (in msec) passed to AIV in AivSubscribeRequest
    }
  },
  "TLEE": {
    "UseInternalTLEE": false            // false: use HUAWEI TLEE implementation
                                        // true: use internal mockup TLEE instead
    "DebuggingMode": false,             // false: disable TLEE debugging features
                                        // true: enable TLEE debugging features
    "FilePath": "debug/"                // path to be used for TLEE debugging file output 
  },
  "V2X" : {
    "NodeTTLsec" : 5,                   // The time to live of a node (vehicle) in seconds based on CPMs.
                                        // If there is no message after that time span, the vehicle
                                        // is considered to be gone. 
    "CheckIntervalSec" : 1              // The interval in seconds how often vehicles should be checked
                                        // for TTL expiries (see above).
  }
}
```

## Updating Message Schema and Auto-Generating Go Structs

**Warning:** *This step is only necessary after modifying existing schemas or adding new schemas. **Don't do this step unless you know that it is really necessary, as it overwrites existing code and may break the existing TAF implementation.***

This step requires `quicktype`. Having node/npm already installed, you can install it using:

```shell
npm install -g quicktype
```

All JSON schemas are located in the folder `res/schemas/`.
By running the command below, corresponding Go structs will be generated into the directory `pkg/message/<namespace>/`. 

```shell
make generate-structs 
```

To remove existing structs, you can use the following command:

```shell
make clean-structs 
```

Again, please note that adding new schemas/structs will require manual code changes in addition to the auto-generation of the structs.
