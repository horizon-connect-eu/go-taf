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

The TAF uses an internal configuration with hardcoded defaults. To change the configuration, you can use a JSON file (template located in `res/taf.json`) and specify the actual file location in the environment variable `TAF_CONFIG`.


## Watch Application for Testing/Debugging

To debug incoming Kafka communication from the perspective of the TAF, this repository provides a helper application that emulates the Kafka topic consumption behavior of the TAF and validates and checks incoming messages. To build this helper, use:

```shell
make build-watch
```

And to run it:

```shell
make run-watch
```
The helper application will now dump any incoming messages on the following topics: "taf", "tch", "aiv", "mbd", "application.ccam".
For each consumed message, the application will do the following:

 * check whether the message is valid JSON
 * check whether the message is valid according to its Schema
 * create a struct according to the type of message (unmarshalling)

## Playback Application for Testing/Debugging

To induce a workload on the TAF, it provides another helper application that creates Kafka messages based on a folder specifying the worklaod.  
To build this helper, use:

```shell
make build-playback
```

And to run it:

```shell
make run-playback
```

The playback helper takes the input from `res/workloads/example` (subject to change, will be configurable) and produces according Kafka messages.

A workload folder consists of two main components: (1) a `script.csv` file that orchestrates the workload and list of JSON files that represent messages to be sent as a record `value`of a Kafka message.
The CSV file uses the following structure:

| Rel. Timestamp since Start (in milliseconds) | Destination Topic Name | JSON File to use as Message Content |
|----------------------------------------------|------------------------|-------------------------------------|

```csv
1000,"taf","001-init.json"
```

In the given example entry, after 1000 ms, the playback components sends a message to the topic "taf" and uses the file content of "001-init.json" as record value. 

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

 * [Trust Assessment Framework Documentation](https://connect.p.lxd-vs.uni-ulm.de/standalone-taf-documentation): User Documentation (WIP; currently UUlm-internal)

