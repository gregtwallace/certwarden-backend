package storage_test

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/domain/orders"
	"certwarden-backend/pkg/helpers_test"
	"fmt"
	"testing"
	"time"
)

func TestPostNewOrder(t *testing.T) {
	testCases := []struct {
		newPayload        orders.NewOrderAcmePayload
		expectedPostOrdID int
		expectedPostErr   error

		expectedGetOrd *orders.Order
		expectedGetErr error
	}{
		// new order
		{
			orders.NewOrderAcmePayload{
				CertId:         33,
				AccountId:      1,
				Status:         "pending",
				KnownRevoked:   false,
				Expires:        new(time.Unix(21241234123, 0)),
				DnsIds:         []string{"abc.example.com"},
				Error:          nil,
				Authorizations: []string{"abc.example.com/authz/123"},
				Finalize:       "abc.example.com/finalize/123",
				Profile:        nil,
				Location:       "abc.example.com/ord/123",
				CreatedAt:      12434213421,
				UpdatedAt:      12434213421,
			},
			208,
			nil,
			&orders.Order{
				ID:             208,
				Certificate:    cert33,
				Location:       "abc.example.com/ord/123",
				Status:         "pending",
				KnownRevoked:   false,
				Error:          nil,
				Expires:        new(time.Unix(21241234123, 0)),
				DnsIdentifiers: []string{"abc.example.com"},
				Authorizations: []string{"abc.example.com/authz/123"},
				Finalize:       "abc.example.com/finalize/123",
				FinalizedKey:   nil,
				CertificateUrl: nil,
				Pem:            nil,
				ValidFrom:      nil,
				ValidTo:        nil,
				ChainRootCN:    nil,
				CreatedAt:      time.Unix(12434213421, 0),
				UpdatedAt:      time.Unix(12434213421, 0),
				Profile:        nil,
				RenewalInfo:    nil,
			},
			nil,
		},
		// duplicate location (case insensitive) // TODO: fix & uncomment
		// {
		// 	orders.NewOrderAcmePayload{
		// 		CertId:         33,
		// 		AccountId:      1,
		// 		Status:         "pending",
		// 		KnownRevoked:   false,
		// 		Expires:        new(time.Unix(12345, 0)),
		// 		DnsIds:         []string{"123.abc.example.com"},
		// 		Error:          nil,
		// 		Authorizations: []string{"abc.example.com/authz/1234"},
		// 		Finalize:       "abc.example.com/finalize/1234",
		// 		Profile:        nil,
		// 		Location:       "https://acme-staging-v02.api.letsencrypt.org/acme/ORDER/red-1/rED-198",
		// 		CreatedAt:      123456,
		// 		UpdatedAt:      123456,
		// 	},
		// 	198,
		// 	nil,
		// 	&ord198,
		// 	nil,
		// },
		// duplicate location (same case)
		{
			orders.NewOrderAcmePayload{
				CertId:         33,
				AccountId:      1,
				Status:         "pending",
				KnownRevoked:   false,
				Expires:        new(time.Unix(12345, 0)),
				DnsIds:         []string{"123.abc.example.com"},
				Error:          nil,
				Authorizations: []string{"abc.example.com/authz/1234"},
				Finalize:       "abc.example.com/finalize/1234",
				Profile:        nil,
				Location:       "https://acme-staging-v02.api.letsencrypt.org/acme/order/red-1/red-198",
				CreatedAt:      123456,
				UpdatedAt:      123456,
			},
			198,
			orders.ErrOrderExists,
			&ord198,
			nil,
		},
		// empty payload (duplicate location)
		{
			orders.NewOrderAcmePayload{
				Location: "https://acme-staging-v02.api.letsencrypt.org/acme/order/red-1/red-198",
			},
			198,
			orders.ErrOrderExists,
			&ord198,
			nil,
		},
		// empty payload (new location)
		{
			orders.NewOrderAcmePayload{
				Location: "https://acme-staging-v02.api.letsencrypt.org/acme/order/red-1/red-209",
			},
			-2,
			helpers_test.NewTestErrorStringComp("FOREIGN KEY constraint failed"),
			nil,
			nil,
		},
		// bare minimum (new location)
		{
			orders.NewOrderAcmePayload{
				CertId:    33,
				AccountId: 1,
				Location:  "https://acme-staging-v02.api.letsencrypt.org/acme/order/red-1/red-209",
			},
			209,
			nil,
			&orders.Order{
				ID:             209,
				Certificate:    cert33,
				Location:       "https://acme-staging-v02.api.letsencrypt.org/acme/order/red-1/red-209",
				Status:         "",
				KnownRevoked:   false,
				Error:          nil,
				Expires:        nil,
				DnsIdentifiers: nil,
				Authorizations: nil,
				Finalize:       "",
				FinalizedKey:   nil,
				CertificateUrl: nil,
				Pem:            nil,
				ValidFrom:      nil,
				ValidTo:        nil,
				ChainRootCN:    nil,
				CreatedAt:      time.Unix(0, 0),
				UpdatedAt:      time.Unix(0, 0),
				Profile:        nil,
				RenewalInfo:    nil,
			},
			nil,
		},
		// new all vals
		{
			orders.NewOrderAcmePayload{
				CertId:       35,
				AccountId:    1,
				Status:       "processing",
				KnownRevoked: true,
				Expires:      new(time.Unix(34555555, 0)),
				DnsIds:       []string{"abc.example.com", "dcd.example.com"},
				Error: new(acme.Error{
					Status: 341,
					Type:   "urn:ietf:params:acme:error:someThing1",
					Detail: "whoops",
				}),
				Authorizations: []string{"abc.example.com/authz/123", "abc.example.com/authz/345"},
				Finalize:       "abc.example.com/finalize/677",
				Profile:        new("a profile"),
				Location:       "abc.example.com/ord/99888",
				CreatedAt:      7778888,
				UpdatedAt:      9998888,
			},
			210,
			nil,
			&orders.Order{
				ID:           210,
				Certificate:  cert35,
				Location:     "abc.example.com/ord/99888",
				Status:       "processing",
				KnownRevoked: true,
				Error: new(acme.Error{
					Status: 341,
					Type:   "urn:ietf:params:acme:error:someThing1",
					Detail: "whoops",
				}),
				Expires:        new(time.Unix(34555555, 0)),
				DnsIdentifiers: []string{"abc.example.com", "dcd.example.com"},
				Authorizations: []string{"abc.example.com/authz/123", "abc.example.com/authz/345"},
				Finalize:       "abc.example.com/finalize/677",
				FinalizedKey:   nil,
				CertificateUrl: nil,
				Pem:            nil,
				ValidFrom:      nil,
				ValidTo:        nil,
				ChainRootCN:    nil,
				CreatedAt:      time.Unix(7778888, 0),
				UpdatedAt:      time.Unix(9998888, 0),
				Profile:        new("a profile"),
				RenewalInfo:    nil,
			},
			nil,
		},
	}

	// create testing service
	store, err := openStorageWithTestData(t, "postneworder")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: post order location: %s", i, tc.newPayload.Location), func(t *testing.T) {
			ordID, err := store.PostNewOrder(&tc.newPayload)
			if !helpers_test.ErrorsIs(err, tc.expectedPostErr) {
				t.Errorf("expected post error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedPostErr), helpers_test.ErrorToVal(err))
			}

			if ordID != tc.expectedPostOrdID {
				t.Fatalf("expected post order id '%d' but got '%d'", tc.expectedPostOrdID, ordID)
			}

			// only compare order if PostNewOrder didn't return error id -2
			if ordID != -2 {
				ord, err := store.GetOneOrder(ordID)
				if !helpers_test.ErrorsIs(err, tc.expectedGetErr) {
					t.Errorf("expected get order error '%s' but got '%s'", helpers_test.ErrorToVal(tc.expectedGetErr), helpers_test.ErrorToVal(err))
				}

				compareOrder(t, ord, tc.expectedGetOrd)
			}

		})
	}
}
