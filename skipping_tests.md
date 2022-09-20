This is a recurring topic in the gopher community.
Please don't use an opt-in-based test skipping mechanism like build tags.
Keep the required knowledge to work with your project to the bare minimum so that a simple basic go command can execute your project's full testing suite:
go test ./...
If you need to skip tests because they are too slow and you are in a hurry, try using testing.Short():
go test -short./...
func Test(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	
	// ... 
}
You can also use environment variables to control what you want to opt out of your current run. (export SKIP_XYTYPE_TEST="true")
SKIP_INTEGRATION_TEST="true" go test ./...
// helper package
func IntegrationTest(tb testing.TB) {
	if _, ok := os.LookupEnv("SKIP_INTEGRATION_TEST"); ok {
		tb.Skip("skipping integration test")
	}
}

// your package's testing package
func Test(t *testing.T) {
	IntegrationTest(t)

	// ...
}
You can also use test helpers that check if the attached resource to your service is available during the test execution where you ping the resource, and if a connection error is returned, then call testing.TB.Skip.
// helper package
func GetXYClient(tb testing.TB) int {
	// load config

	db, err := sql.Open("drv", "dsn")
	if err != nil {
		tb.Skip()
	}
	if err := db.Ping(); err != nil {
		tb.Skip()
	}

	// some struct that uses the attached resource
	// sql open is just an example here
	return 42
}

// your package's testing package
func Test(t *testing.T) {
	client := GetXYClient(t)
	_ = client
	// ...
}
Try to use the go's test caching mechanism as feedback about your system. When you modify a code unrelated to your project's attached resources, such as a domain logic, the testing suite should skip the tests because of the cache. The sign that the go testing tooling can't cache your integration tests in such scenarios might indicate that your system's separation between SRP responsibilities might be violated. (edited) 


