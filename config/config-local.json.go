package config


// TODO: Get the locale list from ConfigHub team. And then Route to the appropriate Region.

var Placements = []byte(`
    [{
        "name": "pii_us:pii",
        "host": "10.100.28.141",
        "port": "8080",
        "locales": ["en_us"]
    },
    {
        "name": "pii_eu:pii",
        "host": "10.101.28.11",
        "port": "8080",
        "locales": ["en_eu"]
    }]
`);

var ValidPIIPlacements = []string{"pii_us:pii", "pii_eu:pii"};

var EmoHostForExistingPIIWebAppRegion = "10.100.28.141:8080";

var EmoApiKeyHostForExistingPIIWebAppRegion = "10.100.28.141:8081";

