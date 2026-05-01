package download

import (
	"certwarden-backend/pkg/output"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

// TODO: actually check pfx creation?

// function to run one pfx test (needed to be special due to the way pfx data is generated)
func onePfxTest(t *testing.T, handler func(w http.ResponseWriter, r *http.Request) *output.JsonError,
	apiKeyHeader *string, apiKeyURL *string, certName string, expectedJsonErr *output.JsonError) {
	r, err := http.NewRequest("GET", "/certwarden/api/v1/download/certificates", nil)
	if err != nil {
		t.Fatal(err)
	}
	// set cert name like the router would
	ctx := r.Context()
	ps := httprouter.Params{{Key: "name", Value: certName}}
	// add api key to url, if specified
	if apiKeyURL != nil {
		ps = append(ps, httprouter.Param{Key: "apiKey", Value: *apiKeyURL})
	}
	ctx = context.WithValue(ctx, httprouter.ParamsKey, ps)
	r = r.WithContext(ctx)

	// add api key header, if specified
	if apiKeyHeader != nil {
		r.Header.Add("x-api-key", *apiKeyHeader)
	}

	// run the test and check the result
	w := httptest.NewRecorder()
	jsonErr := handler(w, r)

	if !errors.Is(jsonErr, expectedJsonErr) {
		t.Errorf("%s: name '%s' with header api-key '%s' and url api-key '%s' returned error '%s' but expected '%s'", getFunctionName(handler),
			certName, stringPointerToVal(apiKeyHeader), stringPointerToVal(apiKeyURL), jsonErr, expectedJsonErr)
	}

	body := w.Body.String()
	if jsonErr != nil && body != "" {
		t.Errorf("%s: name '%s' with header api-key '%s' and url api-key '%s' returned body data but expected none", getFunctionName(handler),
			certName, stringPointerToVal(apiKeyHeader), stringPointerToVal(apiKeyURL))
	} else if jsonErr == nil && body == "" {
		t.Errorf("%s: name '%s' with header api-key '%s' and url api-key '%s' returned empty body but data was expected", getFunctionName(handler),
			certName, stringPointerToVal(apiKeyHeader), stringPointerToVal(apiKeyURL))
	}
}

func TestOutPFXViaHeader(t *testing.T) {
	// create testing service
	app := makeFakeApp(t)
	service, err := NewService(app)
	if err != nil {
		t.Fatal(err)
	}

	// Test: No header provided
	onePfxTest(t, service.DownloadPfxViaHeader, nil, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, nil, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, nil, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, nil, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, nil, nil, "test-e", output.JsonErrUnauthorized)

	// Test: blank/empty apikey provided
	apiKey := ""
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", output.JsonErrUnauthorized)

	// Test: incorrect apikey provided
	apiKey = "something"
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", output.JsonErrUnauthorized)
	apiKey = "something.something"
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", output.JsonErrUnauthorized)

	// Test: just one of the apikeys
	apiKey = "c-abc"
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", output.JsonErrUnauthorized)
	apiKey = "k-123"
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", output.JsonErrUnauthorized)

	// Test: apikey variants
	apiKey = ".k-123"
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", output.JsonErrUnauthorized)
	apiKey = ".c-abc"
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", output.JsonErrUnauthorized)
	apiKey = "k-123."
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", output.JsonErrUnauthorized)
	apiKey = "c-abc."
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", output.JsonErrUnauthorized)
	apiKey = "123.k-123"
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", output.JsonErrUnauthorized)
	apiKey = "k-123.123"
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", output.JsonErrUnauthorized)
	apiKey = "123.c-abc"
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", output.JsonErrUnauthorized)
	apiKey = "c-abc.123"
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", output.JsonErrUnauthorized)

	// Test: correct apikey provided but via url
	apiKey = "c-abc.k-123"
	onePfxTest(t, service.DownloadPfxViaHeader, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	// `b` doesn't have a non-new apikey
	onePfxTest(t, service.DownloadPfxViaHeader, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	// `d` doesnt have a any correct apikey
	onePfxTest(t, service.DownloadPfxViaHeader, nil, &apiKey, "test-e", output.JsonErrUnauthorized)

	// Test: correct new apikey provided but via url
	apiKey = "c-abc-new.k-123-new"
	onePfxTest(t, service.DownloadPfxViaHeader, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaHeader, nil, &apiKey, "test-e", output.JsonErrUnauthorized)

	// Test: correct apikey provided
	apiKey = "c-abc.k-123"
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-a", nil)
	// `b` doesn't have a non-new apikey
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-c", nil)
	// // `d` doesnt have a any correct apikey
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", nil)

	// Test: correct new apikey provided
	apiKey = "c-abc-new.k-123-new"
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-b", nil)
	onePfxTest(t, service.DownloadPfxViaHeader, &apiKey, nil, "test-e", nil)
}

func TestOutPFXViaURL(t *testing.T) {
	// create testing service
	app := makeFakeApp(t)
	service, err := NewService(app)
	if err != nil {
		t.Fatal(err)
	}

	// Test: No header provided
	onePfxTest(t, service.DownloadPfxViaUrl, nil, nil, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, nil, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, nil, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, nil, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, nil, "test-e", output.JsonErrUnauthorized)

	// Test: blank/empty apikey provided
	apiKey := ""
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", output.JsonErrUnauthorized)

	// Test: incorrect apikey provided
	apiKey = "something"
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", output.JsonErrUnauthorized)
	apiKey = "something.something"
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", output.JsonErrUnauthorized)

	// Test: just one of the apikeys
	apiKey = "c-abc"
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", output.JsonErrUnauthorized)
	apiKey = "k-123"
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", output.JsonErrUnauthorized)

	// Test: apikey variants
	apiKey = ".k-123"
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", output.JsonErrUnauthorized)
	apiKey = ".c-abc"
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", output.JsonErrUnauthorized)
	apiKey = "k-123."
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", output.JsonErrUnauthorized)
	apiKey = "c-abc."
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", output.JsonErrUnauthorized)
	apiKey = "123.k-123"
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", output.JsonErrUnauthorized)
	apiKey = "k-123.123"
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", output.JsonErrUnauthorized)
	apiKey = "123.c-abc"
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", output.JsonErrUnauthorized)
	apiKey = "c-abc.123"
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-d", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", output.JsonErrUnauthorized)

	// Test: correct apikey provided
	apiKey = "c-abc.k-123"
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-a", output.JsonErrUnauthorized)
	// `b` doesn't have a non-new apikey
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-c", output.JsonErrUnauthorized)
	// // `d` doesnt have a any correct apikey
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", nil)

	// Test: correct new apikey provided
	apiKey = "c-abc-new.k-123-new"
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-b", output.JsonErrUnauthorized)
	onePfxTest(t, service.DownloadPfxViaUrl, nil, &apiKey, "test-e", nil)

	// Test: correct apikey but api is disabled
	onePfxTest(t, service.DownloadKeyViaUrl, nil, &apiKey, "test-f", output.JsonErrUnauthorized)
}
