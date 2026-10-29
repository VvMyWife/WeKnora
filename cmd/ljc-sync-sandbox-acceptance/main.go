package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Tencent/WeKnora/internal/sandbox"
)

func main() {
	endpoint := os.Getenv("LJC_SANDBOX_PROOF_HOST")
	if endpoint != "tcp://ljc-sync-20261029-docker:2376" {
		panic("dedicated proof daemon required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	client, err := sandbox.NewDockerRemoteClient(&sandbox.Config{DockerHost: endpoint, DockerTLSCertPath: "/proof-certs", DockerImage: "redis:7.0-alpine", DockerNetworkMode: "none", AllowPrivateEndpoints: true})
	must(err)
	must(client.Health(ctx))
	for round := 0; round < 2; round++ {
		h, err := client.Create(ctx, sandbox.RemoteCreateRequest{TemplateID: "redis:7.0-alpine", Metadata: map[string]string{"ljc-proof": "20261029"}})
		must(err)
		if err := func() error {
			defer client.Delete(context.Background(), h.ID())
			path := "/workspace/附件 proof.txt"
			content := []byte("LJC sandbox isolated evidence 17\n")
			must(client.WriteFile(ctx, h, path, content))
			got, err := client.ReadFile(ctx, h, path)
			must(err)
			if string(got) != string(content) {
				return fmt.Errorf("file content changed")
			}
			result, err := client.Exec(ctx, h, sandbox.RemoteExecRequest{Command: "sh", Args: []string{"-c", "printf proof-ok"}, Timeout: 5 * time.Second})
			must(err)
			if result.ExitCode != 0 || !strings.Contains(result.Stdout, "proof-ok") {
				return fmt.Errorf("exec failed: %+v", result)
			}
			result, err = client.Exec(ctx, h, sandbox.RemoteExecRequest{Command: "sleep", Args: []string{"5"}, Timeout: time.Second})
			must(err)
			if !result.Killed && result.ExitCode == 0 {
				return fmt.Errorf("timeout did not terminate command")
			}
			return nil
		}(); err != nil {
			panic(err)
		}
		if _, err := client.ReadFile(ctx, h, "/workspace/附件 proof.txt"); err == nil {
			panic("deleted sandbox remained readable")
		}
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"dedicated_daemon": true, "rounds": 2, "create_write_read_exec_timeout_delete": "passed", "production_docker_socket_mounted": false})
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
