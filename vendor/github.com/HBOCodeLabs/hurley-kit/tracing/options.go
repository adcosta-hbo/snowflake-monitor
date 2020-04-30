package tracing

// Options is used to unmarshall tracing configuration data from a JSON config
// file section. Ideally, this configuration would be present under a top-level
// JSON key called simply "tracing". This allows all services using the tracing
// package to have a standard/uniform configuration section.
//
//    ...
//    "tracing": {
//        "enabled": true,
//        "sampleRate": 0.1,
//        "host": "localhost",
//        "port": 6381
//    },
//    ...
type Options struct {
	Disabled bool     `json:"disabled"`
	Sampler  sampler  `json:"sampler"`
	Reporter reporter `json:"reporter"`
}

// sampler is used to unmarshall the tracing sample config from the JSON
type sampler struct {
	Type  string  `json:"type"`
	Param float64 `json:"param"`
}

// reporter is used to unmarshall the tracing reporter config from the JSON
type reporter struct {
	Host string `json:"agentHost"`
	Port int    `json:"agentPort"`
}
