package ui

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kube "github.com/pierinho13/kubectl-peek/internal/kubernetes"
)

func TestRenderSecretIncludesUsagesAndWarnings(t *testing.T) {
	t.Parallel()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "database-credentials",
			Namespace: "test",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"username": []byte("admin"),
		},
	}

	result := kube.SecretUsageResult{
		Usages: []kube.SecretUsage{
			{
				Kind: "Deployment",
				Name: "backend",
				References: []kube.SecretUsageReference{
					{
						Description: "container environment",
						Path:        "container/backend envFrom",
						Relation:    "uses",
					},
				},
			},
			{
				Kind: "ExternalSecret",
				Name: "database-credentials",
				References: []kube.SecretUsageReference{
					{
						Description: "generated Secret",
						Path:        "spec.target.name",
						Relation:    "produces",
					},
				},
			},
		},
		Warnings: []kube.SecretUsageWarning{
			{
				Resource: "gateways",
				Err:      testError("forbidden"),
			},
		},
	}

	output := RenderSecret(secret, result, true, true)

	expectedParts := []string{
		"Secret:",
		"database-credentials",
		"Used by:",
		"Deployment/backend",
		"uses: container environment (container/backend envFrom)",
		"ExternalSecret/database-credentials",
		"produces: generated Secret (spec.target.name)",
		"Warnings:",
		"gateways: forbidden",
		"username:",
		"admin",
	}

	for _, expected := range expectedParts {
		if !strings.Contains(output, expected) {
			t.Errorf(
				"RenderSecret() output does not contain %q\n\n%s",
				expected,
				output,
			)
		}
	}
}

func TestRenderSecretHidesUsages(t *testing.T) {
	t.Parallel()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "database-credentials",
			Namespace: "test",
		},
		Data: map[string][]byte{
			"username": []byte("admin"),
		},
	}

	result := kube.SecretUsageResult{
		Usages: []kube.SecretUsage{
			{
				Kind: "Deployment",
				Name: "backend",
			},
		},
	}

	output := RenderSecret(secret, result, false, true)

	if strings.Contains(output, "Used by:") {
		t.Fatalf("expected usages to be hidden\n\n%s", output)
	}

	if !strings.Contains(output, "admin") {
		t.Fatalf("expected Secret value to remain visible\n\n%s", output)
	}
}

func TestRenderSecretRedactsValuesWithByteCount(t *testing.T) {
	t.Parallel()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "database-credentials",
			Namespace: "test",
		},
		Data: map[string][]byte{
			"username": []byte("admin"),
		},
	}

	output := RenderSecret(
		secret,
		kube.SecretUsageResult{},
		false,
		false,
	)

	if strings.Contains(output, "admin") {
		t.Fatalf("expected Secret value to be redacted\n\n%s", output)
	}

	if !strings.Contains(output, "<redacted: 5 bytes>") {
		t.Fatalf("expected redacted byte count\n\n%s", output)
	}
}

func TestRenderSecretUsageReferenceWithoutDescription(
	t *testing.T,
) {
	t.Parallel()

	got := renderSecretUsageReference(
		kube.SecretUsageReference{
			Path:     "spec.secretRef.name",
			Relation: "uses",
		},
	)

	want := "uses: spec.secretRef.name"

	if got != want {
		t.Errorf(
			"renderSecretUsageReference() = %q, want %q",
			got,
			want,
		)
	}
}

func TestRenderSecretUsageReferenceIncludesKey(
	t *testing.T,
) {
	t.Parallel()

	got := renderSecretUsageReference(
		kube.SecretUsageReference{
			Description: "environment variable",
			Path:        "container/api env/DATABASE_PASSWORD",
			Key:         "password",
			Relation:    "uses",
		},
	)

	want := "uses: environment variable " +
		"(container/api env/DATABASE_PASSWORD -> password)"

	if got != want {
		t.Errorf(
			"renderSecretUsageReference() = %q, want %q",
			got,
			want,
		)
	}
}

type testError string

func (err testError) Error() string {
	return string(err)
}
