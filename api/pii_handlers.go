package api

import (
	"net/http"
	"io"
	"github.com/gorilla/mux"
	"github.com/golang/glog"
	"io/ioutil"
	"strconv"
	"strings"
	"encoding/json"
	"github.com/piiproxy/config"
)

var placementConfig []PlacementConfig;

type PlacementConfig struct {
	Name    string `json: "name"`
	Host    string `json: "host"`
	Port    string `json: "port"`
	Locales []string `json: "locales"`
}

func (f *PlacementConfig) GetName() string {
	return f.Name;
}
func (f *PlacementConfig) GetHost() string {
	return f.Host;
}
func (f *PlacementConfig) GetPort() string {
	return f.Port;
}

func LoadDataFromJson() {
	err := json.Unmarshal(config.Placements, &placementConfig);
	if err != nil {
		glog.Fatalf("error: %v", err);
		panic(err);
	}
}

/*
- Create a Table

- Users will create their own tables - per client or per client and PII-Type or anything ... (example: pii-bvclient or pii-bvclient-ipaddress etc...)
- If the placement-region is specified, master table will be created based on it, and a facade is created in the other region.
- Otherwise by default, master table is created in EU and the respective Facade is created in US.
*/
func CreateTable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r);
	table := vars["table"];
	placement := r.URL.Query().Get("placement");
	auditComment := r.URL.Query().Get("auditComment");
	body, _ := ioutil.ReadAll(r.Body);
	data := strings.NewReader(string(body[:]));
	contentType := r.Header.Get("Content-Type");

	// If user hasn't specified the placement, then the default is eu.
	if len(placement) == 0 {
		placement = "pii_eu:pii";
	}

	if !contains(config.ValidPIIPlacements, placement) {
		io.WriteString(w, "Not a valid PII Placement");
		return;
	}

	var placementForFacade string;
	for _, value := range config.ValidPIIPlacements {
		if value != placement {
			placementForFacade = value;
		}
	}

	hostForMasterTable := getHostFromPlacement(placementConfig, placement);
	hostForFacade := getHostFromPlacement(placementConfig, placementForFacade);

	// Create master table in the placement Region.
	optionsForMasterTable := "placement:'" + placement + "'";
	if len(auditComment) != 0 {
		optionsForMasterTable = optionsForMasterTable + "&audit=comment:'" + auditComment + "'";
	}
	createTableStatusCode, _ := doPutAction("http://" + hostForMasterTable + "/sor/1/_table/" + table + "?options=" + optionsForMasterTable, data, contentType);
	if (createTableStatusCode != 200) {
		io.WriteString(w, "There was a problem when creating master table. HTTP Response code: " + strconv.Itoa(createTableStatusCode));
		return;
	}
	io.WriteString(w, "Master table created.\n");

	// Create Facade in the other region.
	optionsForFacade := "placement:'" + placementForFacade + "'";
	if len(auditComment) != 0 {
		optionsForFacade = optionsForFacade + "&audit=comment:'" + auditComment + "'";
	}
	createFacadeStatusCode, _ := doPutAction("http://" + hostForFacade + "/sor/1/_facade/" + table + "?options=" + optionsForFacade, data, contentType);
	if (createFacadeStatusCode != 200) {
		io.WriteString(w, "There was a problem when creating facade. HTTP Response code: " + strconv.Itoa(createFacadeStatusCode));
		return;
	}
	io.WriteString(w, "Facade created.\n");

	w.WriteHeader(http.StatusOK);
}

/*
- Store given data

- *** NOTE - This should only write to the Master table in which ever region it exists  ***
*/
func PutPIIData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r);
	table := vars["table"];
	id := vars["id"];
	locale := r.URL.Query().Get("locale");
	auditComment := r.URL.Query().Get("auditComment");
	apiKey := r.URL.Query().Get("APIKey");
	body, _ := ioutil.ReadAll(r.Body);
	data := strings.NewReader(string(body[:]));
	contentType := r.Header.Get("Content-Type");

	// If user hasn't specified the locale, just return.
	if len(locale) == 0 {
		io.WriteString(w, "locale parameter should be specified");
		return;
	}

	// First get the placement for the specified table
	placement := getPlacementForTable(table);
	println(placement);
	var placementForFacade string;
	for _, value := range config.ValidPIIPlacements {
		if value != placement {
			placementForFacade = value;
		}
	}

	hostForMasterTable := getHostFromPlacement(placementConfig, placement);
	hostForFacade := getHostFromPlacement(placementConfig, placementForFacade);

	// Store the Data in master table.
	storeInMasterStatusCode, _ := doPutAction("http://" + hostForMasterTable + "/sor/1/" + table + "/" + id + "/?audit=comment:'" + auditComment + "'", data, contentType);
	if (storeInMasterStatusCode != 200) {
		io.WriteString(w, "There was a problem when storing the document in master table. HTTP Response code: " + strconv.Itoa(storeInMasterStatusCode));
		return;
	}
	io.WriteString(w, "Stored the Data in master table.\n");

	// Store some reference data in the respective Facade in the other Placement.
	dataForFacade := strings.NewReader(`{"Info":"The data for the specified ID exists, but is not in this Placement"}`); //TODO: reword it later.
	// API Key is needed for the facade insert - < delete - created one with the required permissions for time being >.
	storeInFacadeStatusCode, _ := doPutAction("http://" + hostForFacade + "/sor/1/_facade/" + table + "/" + id + "?audit=comment:'" + auditComment + "'&APIKey=" + apiKey, dataForFacade, contentType);
	if (storeInFacadeStatusCode != 200) {
		io.WriteString(w, "There was a problem when storing the document in facade. HTTP Response code: " + strconv.Itoa(storeInFacadeStatusCode));
		return;
	}
	io.WriteString(w, "Stored the Data in facade.\n");

	w.WriteHeader(http.StatusOK);
}

/*
- Return the data
*/
func GetPIIData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r);
	table := vars["table"];
	id := vars["id"];

	// Get data
	// Based on the PII web-app region, first check the corresponding emodb.
	// If there is a facade in this region, then the below url will automatically looks for the facade - TODO : confirm again!
	resp, err := http.Get("http://" + config.EmoHostForExistingPIIWebAppRegion + "/sor/1/" + table + "/" + id);
	if err != nil {
		glog.Error("error from GetPIIData: ", err.Error());
		panic(err);
		return;
	}
	defer resp.Body.Close();
	println("response code:" + strconv.Itoa(resp.StatusCode));
	if (resp.StatusCode != 200) {
		io.WriteString(w, "There was a problem when retrieving the data. HTTP Response code: " + strconv.Itoa(resp.StatusCode));
		return;
	}
	body, _ := ioutil.ReadAll(resp.Body);
	io.WriteString(w, string(body[:]));

	// TODO: if the response body has reference data, then check the other region for the actual data.

	w.WriteHeader(http.StatusOK);
}

func doPutAction(url string, data io.Reader, contentType string) (int, string) {
	glog.Info("PUT URL: " + url);

	req, err := http.NewRequest("PUT", url, data);
	if err != nil {
		glog.Error(err.Error());
		panic(err);
	}
	req.Header.Set("Content-Type", contentType);

	resp, err := http.DefaultClient.Do(req);
	if err != nil {
		glog.Error("error from doPutAction: ", err.Error());
		panic(err);
	}
	defer resp.Body.Close();

	glog.Info("response Status: " + resp.Status);
	body, _ := ioutil.ReadAll(resp.Body);
	glog.Info("response Body: " + string(body));

	return resp.StatusCode, string(body[:]);
}

func getHostFromPlacement(placementConfig []PlacementConfig, placement string) string {
	var host, port string;
	for _, pConfig := range placementConfig {
		if pConfig.GetName() == placement {
			host = pConfig.GetHost();
			port = pConfig.GetPort();
		}
	}
	return host + ":" + port;
}

var tableMetadata []TableMetadata;

type TableMetadata struct {
	Name         string `json: "name"`
	Options      struct {
			     Placement string `json: "placement"`
			     Facades   []struct {
				     Placement string `json: "placement"`
			     } `json: "facades"`
		     } `json: "options"`
	Template     struct {
			     Type   string `json: "type"`
			     Client string `json: "client"`
		     } `json: "template"`
	Availability struct {
			     Placement string `json: "placement"`
			     Facade    bool `json: "facade"`
		     } `json: "availability"`
};

// TODO: do we have any better way than calling "_table?from=<tablename>&limit=1" to find out the placement for the given table? NOT SURE, find out later.....
func getPlacementForTable(table string) string {
	// emodb URL of the corresponding PII web-app region.
	resp, err := http.Get("http://" + config.EmoHostForExistingPIIWebAppRegion + "/sor/1/_table?from=" + table + "&limit=1");
	if err != nil {
		glog.Error("error from getPlacementForTable: ", err.Error());
		panic(err);
	}
	defer resp.Body.Close();
	println("response code:" + strconv.Itoa(resp.StatusCode));
	if (resp.StatusCode != 200) {
		panic("There was a problem when retrieving the data. HTTP Response code: " + strconv.Itoa(resp.StatusCode));
	}
	jsonBody, _ := ioutil.ReadAll(resp.Body);
	err2 := json.Unmarshal([]byte(string(jsonBody[:])), &tableMetadata);
	if err2 != nil {
		glog.Fatalf("error: %v", err2.Error());
		panic(err);
	}

	return tableMetadata[0].Options.Placement;
}

// TODO: delete later.
//func getContentType(h io.Reader) (contentType string) {
//	header, _ := ioutil.ReadAll(h);
//	headerString := string(header[:]);
//	if strings.Contains(headerString, "application/json") {
//		return "application/json";
//	} else if strings.Contains(headerString, "application/x.json-delta") {
//		return "application/x.json-delta";
//	}
//	println("Unsupported content type specified.")
//	os.Exit(1);
//	return;
//}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}