package common

import (
	"github.com/stretchr/testify/assert"
	"testing"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

func TestGitLikeProjectSource_GetDefaultSource(t *testing.T) {
	checkoutFromRemoteUndefinedErr := "checkoutFrom.Remote is not defined in Remotes"
	missingCheckoutFromErr := "there are multiple git remotes but no checkoutFrom information"

	tests := []struct {
		name                 string
		gitLikeProjectSource v1.GitLikeProjectSource
		want1                string
		want2                string
		want3                string
		wantErr              *string
	}{
		{
			name: "only one remote",
			gitLikeProjectSource: v1.GitLikeProjectSource{
				Remotes: map[string]string{
					"origin": "url",
				},
			},
			want1: "origin",
			want2: "url",
			want3: "",
		},
		{
			name: "multiple remotes, checkoutFrom with only branch",
			gitLikeProjectSource: v1.GitLikeProjectSource{
				Remotes: map[string]string{
					"origin": "urlO",
				},
				CheckoutFrom: &v1.CheckoutFrom{Revision: "dev"},
			},
			want1: "origin",
			want2: "urlO",
			want3: "dev",
		},
		{
			name: "multiple remotes, checkoutFrom without revision",
			gitLikeProjectSource: v1.GitLikeProjectSource{
				Remotes: map[string]string{
					"origin":   "urlO",
					"upstream": "urlU",
				},
				CheckoutFrom: &v1.CheckoutFrom{Remote: "upstream"},
			},
			want1: "upstream",
			want2: "urlU",
			want3: "",
		},
		{
			name: "multiple remotes, checkoutFrom with revision",
			gitLikeProjectSource: v1.GitLikeProjectSource{
				Remotes: map[string]string{
					"origin":   "urlO",
					"upstream": "urlU",
				},
				CheckoutFrom: &v1.CheckoutFrom{Remote: "upstream", Revision: "v1"},
			},
			want1: "upstream",
			want2: "urlU",
			want3: "v1",
		},
		{
			name: "multiple remotes, checkoutFrom with unknown remote",
			gitLikeProjectSource: v1.GitLikeProjectSource{
				Remotes: map[string]string{
					"origin":   "urlO",
					"upstream": "urlU",
				},
				CheckoutFrom: &v1.CheckoutFrom{Remote: "non"},
			},
			want1:   "",
			want2:   "",
			want3:   "",
			wantErr: &checkoutFromRemoteUndefinedErr,
		},
		{
			name: "multiple remotes, no checkoutFrom",
			gitLikeProjectSource: v1.GitLikeProjectSource{
				Remotes: map[string]string{
					"origin":   "urlO",
					"upstream": "urlU",
				},
			},
			want1:   "",
			want2:   "",
			want3:   "",
			wantErr: &missingCheckoutFromErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got1, got2, got3, err := GetDefaultSource(tt.gitLikeProjectSource)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestGitLikeProjectSource_GetDefaultSource() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				if got1 != tt.want1 {
					t.Errorf("TestGitLikeProjectSource_GetDefaultSource() error: got1 = %v, want %v", got1, tt.want1)
				}
				if got2 != tt.want2 {
					t.Errorf("TestGitLikeProjectSource_GetDefaultSource() error: got2 = %v, want %v", got2, tt.want2)
				}
				if got3 != tt.want3 {
					t.Errorf("TestGitLikeProjectSource_GetDefaultSource() error: got3 = %v, want %v", got3, tt.want3)
				}
			} else {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestGitLikeProjectSource_GetDefaultSource(): Error message should match")
			}
		})
	}
}

func TestGetProjectSrcType(t *testing.T) {
	projectSrcTypeErr := "unknown project source type"

	tests := []struct {
		name           string
		projectSrc     v1.ProjectSource
		wantErr        *string
		projectSrcType v1.ProjectSourceType
	}{
		{
			name: "Git project",
			projectSrc: v1.ProjectSource{
				Git: &v1.GitProjectSource{},
			},
			projectSrcType: v1.GitProjectSourceType,
		},
		{
			name: "Zip project",
			projectSrc: v1.ProjectSource{
				Zip: &v1.ZipProjectSource{},
			},
			projectSrcType: v1.ZipProjectSourceType,
		},
		{
			name: "Custom project",
			projectSrc: v1.ProjectSource{
				Custom: &v1.CustomProjectSource{},
			},
			projectSrcType: v1.CustomProjectSourceType,
		},
		{
			name:       "Unknown project",
			projectSrc: v1.ProjectSource{},
			wantErr:    &projectSrcTypeErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetProjectSourceType(tt.projectSrc)
			// Unexpected error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("TestGetProjectSrcType() unexpected error: %v, wantErr %v", err, tt.wantErr)
			} else if err == nil && got != tt.projectSrcType {
				t.Errorf("TestGetProjectSrcType() error: project src type mismatch, expected: %v got: %v", tt.projectSrcType, got)
			} else if err != nil {
				assert.Regexp(t, *tt.wantErr, err.Error(), "TestGetProjectSrcType(): Error message should match")
			}
		})
	}

}
