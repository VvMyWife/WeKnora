// A bounded live graph-provider probe. It creates no knowledge or queue records.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Tencent/WeKnora/internal/models/limiter"
	"github.com/redis/go-redis/v9"
)

func main() {
	endpoint := flag.String("endpoint", "http://127.0.0.1:8101/v1/chat/completions", "model endpoint")
	model := flag.String("model", "qwen35-9b", "served model name")
	modelID := flag.String("model-id", "6b0204b6-d092-4ead-9222-7c97c54460fb", "WeKnora graph model ID")
	flag.Parse()
	rdb := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR"), Password: os.Getenv("REDIS_PASSWORD")})
	defer rdb.Close()
	limiter.SetGovernor(limiter.NewRedisLimiter(rdb), 2)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	start := make(chan struct{})
	ready := make(chan struct{}, 9)
	results := make(chan error, 9)
	var active, peak atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 9; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			callCtx, release, err := limiter.AcquireGraph(ctx, *modelID, *model)
			if err != nil {
				results <- err
				return
			}
			defer release()
			ready <- struct{}{}
			select {
			case <-start:
			case <-ctx.Done():
				results <- ctx.Err()
				return
			}
			n := active.Add(1)
			defer active.Add(-1)
			for old := peak.Load(); n > old; old = peak.Load() {
				if peak.CompareAndSwap(old, n) {
					break
				}
			}
			body, _ := json.Marshal(map[string]interface{}{"model": *model, "messages": []map[string]string{{"role": "user", "content": fmt.Sprintf("请逐行列出100条档案数字化工作中的具体检查事项，每条15到25个字，不要省略，不要思考说明。这是并发验证请求%d。", i)}}, "max_tokens": 512, "temperature": 0, "chat_template_kwargs": map[string]bool{"enable_thinking": false}})
			req, _ := http.NewRequestWithContext(callCtx, http.MethodPost, *endpoint, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if key := os.Getenv("MODEL_API_KEY"); key != "" {
				req.Header.Set("Authorization", "Bearer "+key)
			}
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				results <- err
				return
			}
			defer res.Body.Close()
			var response struct {
				Choices []struct{ Message struct{ Content string } }
			}
			if res.StatusCode != 200 {
				b, _ := io.ReadAll(io.LimitReader(res.Body, 500))
				results <- fmt.Errorf("HTTP %d: %s", res.StatusCode, b)
				return
			}
			if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
				results <- err
				return
			}
			if len(response.Choices) == 0 || response.Choices[0].Message.Content == "" {
				results <- fmt.Errorf("empty model response")
				return
			}
			results <- nil
		}(i)
	}
	for i := 0; i < 8; i++ {
		select {
		case <-ready:
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr, "could not acquire eight:", ctx.Err())
			os.Exit(1)
		}
	}
	select {
	case <-ready:
		fmt.Fprintln(os.Stderr, "ninth request admitted before release")
		os.Exit(1)
	case <-time.After(300 * time.Millisecond):
	}
	stats, _, err := limiter.RuntimeStats(ctx)
	if err != nil {
		panic(err)
	}
	metadata, _ := json.Marshal(map[string]interface{}{"phase": "eight-admitted-ninth-waiting", "global_default": 2, "stats": stats})
	fmt.Println(string(metadata))
	close(start)
	wg.Wait()
	close(results)
	failures := 0
	for err := range results {
		if err != nil {
			failures++
			fmt.Fprintln(os.Stderr, err)
		}
	}
	fmt.Printf("completed=9 failed=%d provider_http_peak=%d\n", failures, peak.Load())
	if failures > 0 || peak.Load() != 8 {
		os.Exit(1)
	}
}
