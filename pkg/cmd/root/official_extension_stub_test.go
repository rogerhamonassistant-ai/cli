package root

import (
	"fmt"
	"io"
	"testing"

	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/internal/prompter"
	"github.com/cli/cli/v2/pkg/extensions"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOfficialExtensionStubRun(t *testing.T) {
	ext := &extensions.OfficialExtension{Name: "cool", Owner: "github", Repo: "gh-cool"}

	tests := []struct {
		name             string
		isTTY            bool
		confirmResult    bool
		confirmErr       error
		installErr       error
		dispatchErr      error
		dispatchFound    bool
		args             []string
		wantErr          string
		wantStderr       string
		wantInstalled    bool
		wantDispatched   bool
		wantDispatchArgs []string
	}{
		{
			name:       "non-TTY prints install instructions",
			isTTY:      false,
			wantStderr: "gh extension install github/gh-cool",
		},
		{
			name:             "TTY confirmed installs",
			isTTY:            true,
			confirmResult:    true,
			dispatchFound:    true,
			wantStderr:       "Successfully installed github/gh-cool",
			wantInstalled:    true,
			wantDispatched:   true,
			wantDispatchArgs: []string{"cool"},
		},
		{
			name:             "TTY confirmed forwards args",
			isTTY:            true,
			confirmResult:    true,
			dispatchFound:    true,
			args:             []string{"submit", "--dry-run"},
			wantStderr:       "Successfully installed github/gh-cool",
			wantInstalled:    true,
			wantDispatched:   true,
			wantDispatchArgs: []string{"cool", "submit", "--dry-run"},
		},
		{
			name:          "TTY declined does not install",
			isTTY:         true,
			confirmResult: false,
		},
		{
			name:       "TTY prompt error is propagated",
			isTTY:      true,
			confirmErr: fmt.Errorf("prompt interrupted"),
			wantErr:    "prompt interrupted",
		},
		{
			name:          "TTY install error is propagated",
			isTTY:         true,
			confirmResult: true,
			installErr:    fmt.Errorf("network error"),
			wantErr:       "network error",
			wantInstalled: true,
		},
		{
			name:             "TTY dispatch error is propagated",
			isTTY:            true,
			confirmResult:    true,
			dispatchFound:    true,
			dispatchErr:      fmt.Errorf("dispatch failed"),
			wantErr:          "dispatch failed",
			wantInstalled:    true,
			wantDispatched:   true,
			wantDispatchArgs: []string{"cool"},
		},
		{
			name:             "TTY dispatch not found is propagated",
			isTTY:            true,
			confirmResult:    true,
			dispatchFound:    false,
			wantErr:          "cool was not found after install",
			wantInstalled:    true,
			wantDispatched:   true,
			wantDispatchArgs: []string{"cool"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, stderr := iostreams.Test()
			if tt.isTTY {
				ios.SetStdinTTY(true)
				ios.SetStdoutTTY(true)
				ios.SetStderrTTY(true)
			}

			em := &extensions.ExtensionManagerMock{
				InstallFunc: func(_ ghrepo.Interface, _ string) error {
					return tt.installErr
				},
				DispatchFunc: func(_ []string, _ io.Reader, _ io.Writer, _ io.Writer) (bool, error) {
					return tt.dispatchFound, tt.dispatchErr
				},
			}
			p := &prompter.PrompterMock{
				ConfirmFunc: func(_ string, _ bool) (bool, error) {
					return tt.confirmResult, tt.confirmErr
				},
			}

			err := officialExtensionStubRun(ios, p, em, ext, tt.args)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantStderr != "" {
				assert.Contains(t, stderr.String(), tt.wantStderr)
			}

			if tt.wantInstalled {
				require.NotEmpty(t, em.InstallCalls())
				repo := em.InstallCalls()[0].InterfaceMoqParam
				assert.Equal(t, "github", repo.RepoOwner())
				assert.Equal(t, "gh-cool", repo.RepoName())
				assert.Equal(t, "github.com", repo.RepoHost())
			} else if tt.isTTY && !tt.confirmResult && tt.confirmErr == nil {
				assert.Empty(t, em.InstallCalls())
			}

			if tt.wantDispatched {
				require.NotEmpty(t, em.DispatchCalls())
				assert.Equal(t, tt.wantDispatchArgs, em.DispatchCalls()[0].Args)
			} else {
				assert.Empty(t, em.DispatchCalls())
			}
		})
	}
}

func TestNewCmdOfficialExtensionStub_Properties(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	ext := &extensions.OfficialExtension{Name: "cool", Owner: "github", Repo: "gh-cool"}
	em := &extensions.ExtensionManagerMock{}
	p := &prompter.PrompterMock{}

	cmd := NewCmdOfficialExtensionStub(ios, p, em, ext)

	assert.Equal(t, "cool", cmd.Use)
	assert.True(t, cmd.Hidden)
	assert.Equal(t, "extension", cmd.GroupID)
	assert.True(t, cmd.DisableFlagParsing)
}
