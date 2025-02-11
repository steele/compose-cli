/*
   Copyright 2020 Docker Compose CLI authors

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

package e2e

import (
	"testing"

	"gotest.tools/v3/icmd"

	. "github.com/docker/compose-cli/utils/e2e"
)

func TestLocalComposeExec(t *testing.T) {
	c := NewParallelE2eCLI(t, binDir)

	const projectName = "compose-e2e-exec"

	c.RunDockerCmd("compose", "--project-directory", "fixtures/simple-composefile", "--project-name", projectName, "up", "-d")

	t.Run("exec true", func(t *testing.T) {
		res := c.RunDockerOrExitError("exec", "compose-e2e-exec_simple_1", "/bin/true")
		res.Assert(t, icmd.Expected{ExitCode: 0})
	})

	t.Run("exec false", func(t *testing.T) {
		res := c.RunDockerOrExitError("exec", "compose-e2e-exec_simple_1", "/bin/false")
		res.Assert(t, icmd.Expected{ExitCode: 1})
	})
}
