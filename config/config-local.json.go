package config


// TODO: Get the locale list from ConfigHub team. And then Route to the appropriate Region.

var Placements = []byte(`
    [{
        "name": "pii_us:pii",
        "host": "perf-sept8-useast-emodb-app-062ea20f7045c1213",
        "port": "8080",
        "locales": ["en_us"]
    },
    {
        "name": "pii_eu:pii",
        "host": "perf-sept8-euwest-emodb-app-0ae24aba2458a8986",
        "port": "8080",
        "locales": ["en_eu"]
    }]
`);

var ValidPIIPlacements = []string{"pii_us:pii", "pii_eu:pii"};

var EmoHostForExistingPIIWebAppRegion = "perf-sept8-useast-emodb-app-062ea20f7045c1213:8080";

var EmoApiKeyHostForExistingPIIWebAppRegion = "perf-sept8-useast-emodb-app-062ea20f7045c1213:8081";

