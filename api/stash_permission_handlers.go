package api

import (
	"net/http"
	"io/ioutil"
	"io"

	"github.com/golang/glog"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"strings"
	"github.com/piiproxy/config"
)

const emodbStashReadRole = "test-emodb-stash-bazaar-read-role";

const nexusAccount = "774013277495"; // TODO: this can change, how to deal with it ?
const iamResource = "arn:aws:iam::" + nexusAccount + ":role/" + emodbStashReadRole;

const emodbStashReadPolicyDocument = "{\"Version\": \"2012-10-17\", \"Statement\": {\"Effect\": \"Allow\", \"Action\": \"sts:AssumeRole\", \"Resource\": \"" + iamResource + "\"}}";

// TODO: get from some where.
const adminAPIKey = "local_admin";

/*
- Validates the API key.
- Adds the policy document to the given role.
- Updates the APIKey with the IAM Role.
 */
func AssignReadAccess(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role");
	key := r.URL.Query().Get("key");

	// Validate APIKey
	// TODO: configure hostname and others
	validateApiKeyUrl := "http://" + config.EmoApiKeyHostForExistingPIIWebAppRegion + "/tasks/api-key?action=view&APIKey=" + adminAPIKey + "&key=" + key;
	_, response := postURL(validateApiKeyUrl);
	if strings.Contains(response, "Unknown key") {
		io.WriteString(w, "Unknown API Key");
		return;
	}

	// Add the policy document
	svc := iam.New(session.New()); // TODO: have one ready when the server starts
	params := &iam.PutRolePolicyInput{
		RoleName:       aws.String(role),
		PolicyName:     aws.String("allow-stash-read"),
		PolicyDocument: aws.String(emodbStashReadPolicyDocument),
	};
	resp, err := svc.PutRolePolicy(params);
	if err != nil {
		glog.Error(err.Error());
		panic(err);
		return;
	}
	glog.Info(resp.GoString());

	// Update APIKey
	attributeKey := "read_stash";
	addRoleUrl := "http://" + config.EmoApiKeyHostForExistingPIIWebAppRegion + "/tasks/api-key?action=update_attribute&APIKey=" + adminAPIKey + "&key=" + key + "&attributeKey=" + attributeKey + "&addAttributeValue=" + role;
	statusCode, _ := postURL(addRoleUrl);
	if (statusCode != 200) {
		io.WriteString(w, "There was a problem when updating the API key. Returned Response code: " + string(statusCode));
		return;
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Success");
}

/*
- Validates the API key.
- Deletes the policy document from the given role.
- Updates the APIKey.
 */
func RemoveReadAccess(w http.ResponseWriter, r *http.Request) {
	//role := r.URL.Path["/stash-read/remove/":];
	role := r.URL.Query().Get("role");
	key := r.URL.Query().Get("key");

	// Validate APIKey
	// TODO: configure hostname and others
	validateApiKeyUrl := "http://" + config.EmoApiKeyHostForExistingPIIWebAppRegion + "/tasks/api-key?action=view&APIKey=" + adminAPIKey + "&key=" + key;
	_, response := postURL(validateApiKeyUrl);
	if strings.Contains(response, "Unknown key") {
		io.WriteString(w, "Unknown API Key");
		return;
	}

	// Delete the policy document
	svc := iam.New(session.New()); // TODO: have one ready when the server starts
	params := &iam.DeleteRolePolicyInput{
		RoleName:   aws.String(role),
		PolicyName: aws.String("allow-stash-read"),
	};
	resp, err := svc.DeleteRolePolicy(params);
	if err != nil {
		glog.Error(err.Error());
		panic(err);
		return
	}
	glog.Info(resp);

	// Update APIKey
	attributeKey := "read_stash";
	removeRoleUrl := "http://" + config.EmoApiKeyHostForExistingPIIWebAppRegion + "/tasks/api-key?action=update_attribute&APIKey=" + adminAPIKey + "&key=" + key + "&attributeKey=" + attributeKey + "&removeAttributeValue=" + role;
	statusCode, _ := postURL(removeRoleUrl);
	if (statusCode != 200) {
		io.WriteString(w, "There was a problem when updating the API key. Returned Response code: " + string(statusCode));
		return;
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Success");
}


// TODO: We may not need this - DELETE later
/*
- Gets the IAM Role information.
 */
func GetRoleInfo(w http.ResponseWriter, r *http.Request) {
	//role := r.URL.Path[len("/role/"):]
	role := r.URL.Query().Get("role");

	params := &iam.GetRolePolicyInput{
		RoleName:       aws.String(role),
		PolicyName:     aws.String("allow-stash-read"),
	};
	svc := iam.New(session.New());
	resp, err := svc.GetRolePolicy(params);

	if err != nil {
		glog.Error(err.Error());
		panic(err);
		return
	}
	glog.Info(resp);

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp.GoString()));
}

func postURL(url string) (int, string) {
	req, err := http.NewRequest("POST", url, nil);
	client := &http.Client{};
	resp, err := client.Do(req);
	if err != nil {
		panic(err);
	}
	defer resp.Body.Close();

	glog.Info("response Status: " + resp.Status);
	body, _ := ioutil.ReadAll(resp.Body);
	glog.Info("response Body: " + string(body));

	return resp.StatusCode, string(body[:]);
}