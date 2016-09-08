#PIIProxy

- IAM permissions for reading the PII stash tables.
- Submission and Lookup API for PII data.

##Local

1. Clone the repo
2. Install Godep, download dependencies:
```
$ go get github.com/tools/godep
$ godep get
```
3. Compile and run the server in the root directory:
```
$ cd server; go run emo-piiproxy-server.go
```
4. Check [http://localhost:8888/](http://localhost:8888/)

## Examples: (**DELETE LATER**)
IAM Permission API:

- All the API test run commands are here: [EMO-5982:comment](https://bits.bazaarvoice.com/jira/browse/EMO-5982?focusedCommentId=1754422&page=com.atlassian.jira.plugin.system.issuetabpanels:comment-tabpanel#comment-1754422)

Submission API:
```
sujith.vaddi$ curl -s -XPUT -H "Content-Type: application/json" "http://localhost:8888/pii/1/_table/table2?placement=pii_us:pii&auditComment=piiproxy" --data-binary '{"type":"pii", "client":"testcustomer"}'
Master table created.
Facade created.

sujith.vaddi$ curl -s -XPUT -H "Content-Type: application/json" "http://localhost:8888/pii/1/table2/id1?auditComment=commentsujith&locale='en-us'&APIKey=7emdglvwsy0mfydyxnqafwi490ykanf4amqlzte0njyq7t8k" --data-binary '{"author":"Sujith","title":"Best Ever!","rating":5}'
Stored the Data in master table.
Stored the Data in facade.
```

Lookup API:
```
sujith.vaddi$ curl -s "http://localhost:8888/pii/1/table2/id1"
{"rating":5,"title":"Best Ever!","author":"Sujith","type":"pii","client":"testcustomer","~id":"id1","~table":"table2","~version":1,"~signature":"8c1b4581b345a223dc642cc21f70cc35","~deleted":false,"~firstUpdateAt":"2016-09-07T16:48:40.440Z","~lastUpdateAt":"2016-09-07T16:48:40.440Z"}
```
