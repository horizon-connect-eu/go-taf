module github.com/vs-uulm/go-taf

go 1.22.1

require github.com/vs-uulm/taf-tlee-interface v0.0.0-0ebc6b71f14647f8a7bebeb7b5142ad0a4e7b178
require github.com/vs-uulm/go-subjectivelogic v0.0.0-20240314142756-d2653fbfb4de

replace (
	github.com/vs-uulm/go-subjectivelogic => ../go-subjectivelogic
	github.com/vs-uulm/taf-tlee-interface => ../tlee-interface
)
