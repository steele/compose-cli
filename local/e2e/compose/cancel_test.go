// +build !windows

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
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/icmd"

	. "github.com/docker/compose-cli/utils/e2e"
)

func TestComposeCancel(t *testing.T) {
	c := NewParallelE2eCLI(t, binDir)
	s := NewMetricsServer(c.MetricsSocket())
	s.Start()
	defer s.Stop()

	started := false

	for i := 0; i < 30; i++ {
		c.RunDockerCmd("help", "ps")
		if len(s.GetUsage()) > 0 {
			started = true
			fmt.Printf("    [%s] Server up in %d ms\n", t.Name(), i*100)
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.Assert(t, started, "Metrics mock server not available after 3 secs")

	t.Run("metrics on cancel Compose build", func(t *testing.T) {
		s.ResetUsage()

		c.RunDockerCmd("compose", "ls")
		buildProjectPath := "../compose/fixtures/build-infinite/docker-compose.yml"

		// require a separate groupID from the process running tests, in order to simulate ctrl+C from a terminal.
		// sending kill signal
		cmd, stdout, stderr, err := StartWithNewGroupID(c.NewDockerCmd("compose", "-f", buildProjectPath, "build", "--progress", "plain"))
		assert.NilError(t, err)

		c.WaitForCondition(func() (bool, string) {
			out := stdout.String()
			errors := stderr.String()
			return strings.Contains(out, "RUN sleep infinity"), fmt.Sprintf("'RUN sleep infinity' not found in : \n%s\nStderr: \n%s\n", out, errors)
		}, 30*time.Second, 1*time.Second)

		err = syscall.Kill(-cmd.Process.Pid, syscall.SIGINT) // simulate Ctrl-C : send signal to processGroup, children will have same groupId by default

		assert.NilError(t, err)
		c.WaitForCondition(func() (bool, string) {
			out := stdout.String()
			errors := stderr.String()
			return strings.Contains(out, "CANCELED"), fmt.Sprintf("'CANCELED' not found in : \n%s\nStderr: \n%s\n", out, errors)
		}, 10*time.Second, 1*time.Second)

		usage := s.GetUsage()
		assert.DeepEqual(t, []string{
			`{"command":"compose ls","context":"moby","source":"cli","status":"success"}`,
			`{"command":"compose build","context":"moby","source":"cli","status":"canceled"}`,
		}, usage)
	})
}

func StartWithNewGroupID(command icmd.Cmd) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer, error) {
	cmd := exec.Command(command.Command[0], command.Command[1:]...)
	cmd.Env = command.Env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Start()
	return cmd, &stdout, &stderr, err
}
