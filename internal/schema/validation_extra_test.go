package schema

import "testing"

func TestValidationExtra_WebSocketValid(t *testing.T) {
	cfg := &SmokeConfig{
		Version: 1,
		Project: "ws-app",
		Tests: []Test{
			{
				Name: "ws-echo",
				Expect: Expect{
					WebSocket: &WebSocketCheck{
						URL:            "ws://localhost:8080/echo",
						Send:           `{"action":"ping"}`,
						ExpectContains: "pong",
					},
				},
			},
		},
	}
	if err := Validate(cfg); err != nil {
		t.Errorf("unexpected error for valid websocket config: %v", err)
	}
}

func TestValidationExtra_GraphQLValid(t *testing.T) {
	cfg := &SmokeConfig{
		Version: 1,
		Project: "gql-app",
		Tests: []Test{
			{
				Name: "gql-introspect",
				Expect: Expect{
					GraphQL: &GraphQLCheck{
						URL:   "http://localhost:4000/graphql",
						Query: `{ __schema { types { name } } }`,
					},
				},
			},
		},
	}
	if err := Validate(cfg); err != nil {
		t.Errorf("unexpected error for valid graphql config: %v", err)
	}
}

func TestValidationExtra_CredentialCheckSources(t *testing.T) {
	sources := []struct {
		name   string
		source string
	}{
		{"env", "env"},
		{"file", "file"},
		{"exec", "exec"},
	}
	for _, src := range sources {
		t.Run(src.name, func(t *testing.T) {
			cfg := &SmokeConfig{
				Version: 1,
				Project: "cred-app",
				Tests: []Test{
					{
						Name: "check-" + src.name,
						Expect: Expect{
							Credential: &CredentialCheck{
								Source:   src.source,
								Name:     "MY_SECRET",
								Contains: "token",
							},
						},
					},
				},
			}
			if err := Validate(cfg); err != nil {
				t.Errorf("unexpected error for credential_check source=%s: %v", src.source, err)
			}
		})
	}
}

func TestValidationExtra_S3BucketCustomEndpoint(t *testing.T) {
	cfg := &SmokeConfig{
		Version: 1,
		Project: "s3-app",
		Tests: []Test{
			{
				Name: "minio-bucket",
				Expect: Expect{
					S3Bucket: &S3BucketCheck{
						Bucket:   "test-bucket",
						Region:   "us-east-1",
						Endpoint: "http://localhost:9000",
					},
				},
			},
		},
	}
	if err := Validate(cfg); err != nil {
		t.Errorf("unexpected error for s3_bucket with custom endpoint: %v", err)
	}
}
