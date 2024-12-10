# Standalone TAF Prototype

## Release v0.2.8 (2024-12-10)

 * added experimental web UI for exploring TAF-internal states
   * overview of sessions
   * overview of trust model instances
   * detailed view of trust model instances and their history
 * added support for the TAF to handle errors returned by the TLEE


## Release v0.2.7 (2024-11-22)

* restructured the folders for trust models to allow for multiple templates of the same group
* added automatic signing hashes for trust model templates 
* added support for the following message types
	* `TAS_TMT_DISCOVER`
	* `TAS_TMT_OFFER`


## Release v0.2.6 (2024-11-04)

* added support for queries based on the  Trust Assessment Query Interface (TAQI) which allows to query for instantiated trust models and their ATLs
* added support for the following message types
	* `TAQI_QUERY`
	* `TAQI_RESULT`


## Release v0.2.5 (2024-10-30)

* included a new version of the TLEE (`887019e9050e9a50f8746526452d5089fd9a2da1`)
  * configurable debugging behavior and debugging output
  * increased robustness and improved error handling 
  * unified logging with TAF
* fixed bug in which no TAS_NOTIFY is sent after spawning a new trust model instance
* changed behavior of parsing TCH_NOTIFY messages to allow for different formats allowed in the message specification
* fixed bug in `IMA_STANDALONE@0.0.1` trust model that caused incorrect trust source quantifications for TCH 


## Release v0.2.4 (2024-10-29)

* fixed a bug when handling atomic trust opinion updates in the IMA use case when the TCH NOTIFY message also includes component information


## Release v0.2.3 (2024-10-10)

* fixed handling of atomic trust opinion updates
* improved support for concurrent trust models with identical evidence types coming from different trust sources


## Release v0.2.2 (2024-10-09)

* fixed behavior of TAS_NOTIFY according to the TAS subscription specification
  * Notifications now always include the full set of propositions (instead of only the changed propositions) in case the subscription trigger fires for a trust model instance after it has been modified. 


## Release v0.2.1 (2024-10-01)

* reworked support for handling different trust sources
	* AIV: one separate subscription for each session
	* MBD: single subscription for all sessions
	* TCH: subscription-less
* upgraded internal TLEE to support IMA_STANDALONE trust models for debugging


## Release v0.2.0 (2024-09-30)

* added support for dynamic trust models
  * added support for dynamically spawned trust model instances
  * added support for trust models with dynamically changing topologies
* updated internal TMT/TMI API
* added CPM-based dynamic V2X observer as event source
* added support for TCH as evidence provider
* added support for MBD as evidence provider
* added support for the following message types
	* `MBD_SUBSCRIBE_REQUEST`
	* `MBD_SUBSCRIBE_RESPONSE`
	* `MBD_UNSUBSCRIBE_REQUEST`
	* `MBD_UNSUBSCRIBE_RESPONSE`
	* `MBD_NOTIFY`
	* `TCH_NOTIFY`
	* `V2X_CPM`
* added trust models
  * `IMA_STANDALONE@0.0.1`: Intersection Movement Assist (Standalone Variant) Trust Model
* version of TLEE included: `aa0aa59b4b4362e54430f437607ed5ac7a96a54e`
* version of crypto library included: v1.2


## Release v0.1.1 (2024-09-10)
 
 * fixed integration of the `VCM@0.0.1` tust model


## Release v0.1.0 (2024-08-09)
 
 * initial CONNECT-internal release of the standalone TAF prototype
 * added support for static trust models
 * added support for AIV as evidence provider
 * added support for the following message types
	 * `TAS_INIT_REQUEST`
	 * `TAS_INIT_RESPONSE`
	 * `TAS_TEARDOWN_REQUEST`
	 * `TAS_TEARDOWN_RESPONSE`
	 * `TAS_TA_REQUEST ("TAR")`
	 * `TAS_TA_RESPONSE`
	 * `TAS_SUBSCRIBE_REQUEST`
	 * `TAS_SUBSCRIBE_RESPONSE`
	 * `TAS_UNSUBSCRIBE_REQUEST`
	 * `TAS_UNSUBSCRIBE_RESPONSE`
	 * `TAS_NOTIFY`
	 * `AIV_REQUEST`
	 * `AIV_RESPONSE`
	 * `AIV_SUBSCRIBE_REQUEST`
	 * `AIV_SUBSCRIBE_RESPONSE`
	 * `AIV_UNSUBSCRIBE_REQUEST`
	 * `AIV_UNSUBSCRIBE_RESPONSE`
	 * `AIV_NOTIFY`
 * added trust models 
   * `BRUSSELS@0.0.1`: backport of the Brussels demo
   * `VCM@0.0.1`: vehicle computer migration (DENSO use case)
 * added trust decision engine logic
   * projected probability
 * version of TLEE included: aa0aa59b4b4362e54430f437607ed5ac7a96a54e
 * version of crypto library included: v1.2
