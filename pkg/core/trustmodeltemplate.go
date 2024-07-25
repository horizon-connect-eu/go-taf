package core

type TrustModelTemplate interface {
	TemplateName() string
	Version() string
	Spawn(params map[string]string, context TafContext, channels TafChannels) TrustModelInstance
	EvidenceSources() []Evidence
}
