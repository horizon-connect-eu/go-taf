# Trust Assessment Framework

This repository provides the latest prototype of the standalone Trust Assessment Framework.

## Gettting Started

First, clone this repository:
```shell
git clone git@connect.informatik.uni-ulm.de:coordination/go-taf.git
```

Also clone the following internal dependencies into a shared common folder:
```shell
git clone git@connect.informatik.uni-ulm.de:coordination/tlee-implementation.git
git clone git@connect.informatik.uni-ulm.de:coordination/crypto-library-interface.git
```

The resulting folder structure should look like this:
```
├── crypto-library-interface
├── go-taf
└── tlee-implementation
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

## Configuration

The TAF uses an internal configuration with hardcoded defaults. To change the configuration, you can use a JSON file (template located in `res/taf.json`) and specify the actual file location in the environment variable `TAF_CONFIG`. The following options can be configured. Missing options are implicitly using their defined default values.

```json
{
  "ChanBufSize": 10000,
  "Identifier": "taf",
  "Communication": {
    "Handler": "kafka-based",           // either 'kafka-based' or 'file-based'
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
    "Enabled": false,                   // whether the crypto library should be used or not
    "KeyFolder": "res/cert/",           // path to key folder that is passed to crypto library
    "IgnoreVerificationResults": true   // true: discard messages that failed to verify
                                        // false: process messages that failed to verify
                                        //        (a warning will be logged to console)
  },
  "Debug": {
    "FixedSessionID": "",               // if provided, this fixed value is used by the TAM
                                        // instead of a random UUID-based session id
    "FixedSubscriptionID": "",          // if provided, this fixed value is used by the TAM
                                        // instead of a random UUID-based subscription id
    "FixedRequestID": "1235"            // if provided, this fixed request id is used by the
                                        // trust source manager instead of a random UUID-based id
  },
  "Evidence": {
    "AIV": {
      "CheckInterval": 1000             // check interval passed to AIV in AivSubscribeRequest
    }
  },
  "TAM": {
    "TrustModelInstanceShards": 1       // number of workers used by TAM
  },
  "TLEE": {
    "UseInternalTLEE": true             // use internal mock TLEE or actual TLEE implementation
  },
  "V2X": {
    "NodeTTLsec": 5,                    // TTL of observed V2X nodes
    "CheckIntervalSec": 1               // interval to scan for expired nodes
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


## See Also

 * [Tools for Standalone TAF](https://connect.informatik.uni-ulm.de/coordination/go-taf-tools): `playback` and `watch` tools for TAF development and testing
 * [Trust Assessment Framework Documentation](https://connect.p.lxd-vs.uni-ulm.de/standalone-taf-documentation): User Documentation (WIP; currently UUlm-internal)

