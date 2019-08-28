/*
Copyright 2019 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/runner/runcontext"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestDeploymentCheckStatus(t *testing.T) {
	rolloutCmd := "kubectl --context kubecontext rollout status deployment dep --namespace test --watch=false"
	tests := []struct {
		description    string
		command        util.Command
		expectedErr    string
		expectedReason string
		complete       bool
	}{
		{
			description: "rollout status success",
			command: testutil.NewFakeCmd(t).
				WithRunOut(rolloutCmd, "deployment dep successfully rolled out"),
			expectedReason: "deployment dep successfully rolled out",
			complete:       true,
		}, {
			description: "resource not complete",
			command: testutil.NewFakeCmd(t).
				WithRunOut(rolloutCmd, "Waiting for replicas to be available"),
			expectedReason: "Waiting for replicas to be available",
		}, {
			description: "no output",
			command: testutil.NewFakeCmd(t).
				WithRunOut(rolloutCmd, ""),
		}, {
			description: "rollout status error",
			command: testutil.NewFakeCmd(t).
				WithRunOutErr(rolloutCmd, "", fmt.Errorf("error")),
			expectedErr: "error",
			complete:    true,
		},
		{
			description: "rollout kubectl client connection error",
			command: testutil.NewFakeCmd(t).
				WithRunOutErr(rolloutCmd, "", fmt.Errorf("Unable to connect to the server")),
			expectedReason: "Unable to connect to the server",
		},
	}

	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			t.Override(&util.DefaultExecCommand, test.command)
			r := Deployment{ResourceObj: &ResourceObj{namespace: "test", name: "dep"}}
			runCtx := &runcontext.RunContext{
				KubeContext: "kubecontext",
			}

			r.CheckStatus(context.Background(), runCtx)
			t.CheckDeepEqual(test.complete, r.IsStatusCheckComplete())
			if test.expectedErr != "" {
				t.CheckErrorContains(test.expectedErr, r.Status().Error())
			} else {
				t.CheckDeepEqual(r.status.reason, test.expectedReason)
			}
		})
	}
}
