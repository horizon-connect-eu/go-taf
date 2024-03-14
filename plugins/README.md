# Plugin Docs

## Why?

With plugins, you can override the default behaviour of the TAF.

## How?

Create a go package in this directory (e.g. as `./tam/your_plugin/plugin_name.go`).
In the go file you created, write your functionality as a function (not named `main`).
Create an init function that registers your newly created function.
For this, call one of the `Register[...]` functions with your newly created function as an argument.

### Available `Register[...]` functions

- `RegisterUpdateResultFunc(string, tam.ResultsUpdater)`: Register a function for updating the Results given a State.
- pending.
