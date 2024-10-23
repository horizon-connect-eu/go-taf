// Package web provides utilities for embedded web interfaces to introspect the TAF and other components.
package web

/*
REST API: Implemented Routes:

/sessions

/events
/events?cursor=cursor-id
/events/latest
/events/all

/tmis
/tmis/:clientid/:sessionid/:tmt-id/:tmi-id/:version
/tmis/:clientid/:sessionid/:tmt-id/:tmi-id/ => /tmis/:clientid/:sessionid/:tmt-id/:tmi-id/latest

/info

/trustsources

/trustmodels
/trustmodels/:identifier

*/
