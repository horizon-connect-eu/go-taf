# Standalone TAF Prototype

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
