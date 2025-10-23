# Monad Dashboard - 실제 데이터 업그레이드 가이드

현재 대시보드는 추정 데이터를 사용하고 있습니다. 실제 Monad 메트릭으로 업그레이드하는 방법입니다.

---

## 현재 상태 vs 목표

### 현재 (추정 데이터):
```go
// waterfall_metrics.go
rpcReceived := txCount * 7 / 10      // 70% 가정
sigFailed := totalIngress / 20        // 5% 가정
parallelSuccess := selected * 85 / 100 // 85% 가정
```

### 목표 (실제 데이터):
```go
// 실제 Monad 메트릭에서 읽어옴
rpcReceived := monadMetrics.TxPool.RPCReceived
sigFailed := monadMetrics.TxPool.SignatureFailed
parallelSuccess := monadMetrics.Executor.ParallelSuccess
```

---

## 업그레이드 방법

### **방법 1: Prometheus Metrics 연동 (추천)**

#### 1.1 Monad Prometheus 엔드포인트 찾기

Monad 노드 실행 시 메트릭 서버 주소 확인:
```bash
# Monad 프로세스에서 metrics 포트 찾기
ps aux | grep monad
netstat -tlnp | grep monad

# 일반적인 Prometheus 엔드포인트:
# http://localhost:9090/metrics
# http://localhost:6060/metrics
# http://localhost:2112/metrics
```

#### 1.2 Go Prometheus Client 추가

```go
// backend/go.mod에 추가
require (
    github.com/prometheus/client_golang v1.17.0
    github.com/prometheus/common v0.45.0
)
```

#### 1.3 Metrics Scraper 구현

```go
// backend/prometheus_collector.go
package main

import (
    "fmt"
    "io"
    "net/http"
    "strings"

    dto "github.com/prometheus/client_model/go"
    "github.com/prometheus/common/expfmt"
)

type PrometheusCollector struct {
    metricsURL string
}

func NewPrometheusCollector(url string) *PrometheusCollector {
    return &PrometheusCollector{
        metricsURL: url,
    }
}

// ScrapeMetrics fetches and parses Prometheus metrics
func (p *PrometheusCollector) ScrapeMetrics() (map[string]float64, error) {
    resp, err := http.Get(p.metricsURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    parser := expfmt.TextParser{}
    metricFamilies, err := parser.TextToMetricFamilies(strings.NewReader(string(body)))
    if err != nil {
        return nil, err
    }

    metrics := make(map[string]float64)
    for name, mf := range metricFamilies {
        for _, m := range mf.Metric {
            value := getMetricValue(m)
            metrics[name] = value
        }
    }

    return metrics, nil
}

func getMetricValue(m *dto.Metric) float64 {
    if m.Gauge != nil {
        return m.Gauge.GetValue()
    }
    if m.Counter != nil {
        return m.Counter.GetValue()
    }
    if m.Histogram != nil {
        return m.Histogram.GetSampleSum()
    }
    return 0
}

// GetWaterfallMetrics extracts waterfall-specific metrics
func (p *PrometheusCollector) GetWaterfallMetrics() (*WaterfallRealMetrics, error) {
    metrics, err := p.ScrapeMetrics()
    if err != nil {
        return nil, err
    }

    return &WaterfallRealMetrics{
        // Network ingress
        RPCReceived:    int64(metrics["monad_txpool_rpc_received_total"]),
        P2PReceived:    int64(metrics["monad_txpool_p2p_received_total"]),

        // Verification failures
        SigFailed:      int64(metrics["monad_txpool_sig_failed_total"]),
        NonceFailed:    int64(metrics["monad_txpool_nonce_failed_total"]),
        BalanceFailed:  int64(metrics["monad_txpool_balance_failed_total"]),

        // Pool management
        PendingTxs:     int64(metrics["monad_txpool_pending_txs"]),
        QueuedTxs:      int64(metrics["monad_txpool_queued_txs"]),
        FeeTooLow:      int64(metrics["monad_txpool_fee_too_low_total"]),

        // Execution
        ParallelSuccess:     int64(metrics["monad_executor_parallel_success_total"]),
        SequentialFallback:  int64(metrics["monad_executor_sequential_fallback_total"]),
        ExecFailed:          int64(metrics["monad_executor_exec_failed_total"]),
        StateReads:          int64(metrics["monad_executor_state_reads_total"]),
        StateWrites:         int64(metrics["monad_executor_state_writes_total"]),

        // Block finalization
        BlocksProposed:  int64(metrics["monad_consensus_proposed_total"]),
        BlocksFinalized: int64(metrics["monad_consensus_finalized_total"]),
    }, nil
}

type WaterfallRealMetrics struct {
    RPCReceived         int64
    P2PReceived         int64
    SigFailed           int64
    NonceFailed         int64
    BalanceFailed       int64
    PendingTxs          int64
    QueuedTxs           int64
    FeeTooLow           int64
    ParallelSuccess     int64
    SequentialFallback  int64
    ExecFailed          int64
    StateReads          int64
    StateWrites         int64
    BlocksProposed      int64
    BlocksFinalized     int64
}
```

#### 1.4 기존 코드 업데이트

```go
// backend/waterfall_metrics.go 수정

var prometheusCollector *PrometheusCollector

func InitPrometheusCollector(metricsURL string) {
    prometheusCollector = NewPrometheusCollector(metricsURL)
}

func GenerateWaterfallFromRealMetrics() map[string]interface{} {
    if prometheusCollector == nil {
        return GenerateWaterfallFromSubscriber() // Fallback
    }

    realMetrics, err := prometheusCollector.GetWaterfallMetrics()
    if err != nil {
        log.Printf("Failed to get real metrics: %v", err)
        return GenerateWaterfallFromSubscriber() // Fallback
    }

    return map[string]interface{}{
        "in": map[string]interface{}{
            "rpc":    realMetrics.RPCReceived,     // 실제 값!
            "p2p":    realMetrics.P2PReceived,     // 실제 값!
            "gossip": realMetrics.P2PReceived,
        },
        "out": map[string]interface{}{
            "verify_failed":      realMetrics.SigFailed,         // 실제 값!
            "nonce_failed":       realMetrics.NonceFailed,       // 실제 값!
            "balance_failed":     realMetrics.BalanceFailed,     // 실제 값!
            "pool_fee_dropped":   realMetrics.FeeTooLow,         // 실제 값!
            "pool_full":          int64(0),
            "exec_parallel":      realMetrics.ParallelSuccess,   // 실제 값!
            "exec_sequential":    realMetrics.SequentialFallback, // 실제 값!
            "exec_failed":        realMetrics.ExecFailed,        // 실제 값!
            "state_reads":        realMetrics.StateReads,        // 실제 값!
            "state_writes":       realMetrics.StateWrites,       // 실제 값!
            "block_proposed":     realMetrics.BlocksProposed,    // 실제 값!
            "block_finalized":    realMetrics.BlocksFinalized,   // 실제 값!
        },
    }
}
```

#### 1.5 Main에서 초기화

```go
// backend/main.go에 추가

func main() {
    // ... 기존 코드 ...

    // Prometheus 메트릭 초기화
    metricsURL := os.Getenv("MONAD_METRICS_URL")
    if metricsURL == "" {
        metricsURL = "http://localhost:9090/metrics" // 기본값
    }
    InitPrometheusCollector(metricsURL)
    log.Printf("Prometheus collector initialized: %s", metricsURL)

    // ... 기존 코드 ...
}
```

---

### **방법 2: Event Ring 직접 읽기 (고급)**

Event Ring은 C/Rust로 구현된 고성능 공유 메모리 링버퍼입니다.

#### 2.1 Rust FFI 바인딩 생성

```rust
// monad-bft/monad-event-ring-go-bindings/src/lib.rs
use monad_event_ring::{EventRingReader, ExecEventDecoder};

#[no_mangle]
pub extern "C" fn monad_event_ring_open(path: *const c_char) -> *mut EventRingReader {
    // Event ring 열기
}

#[no_mangle]
pub extern "C" fn monad_event_ring_read_next(reader: *mut EventRingReader) -> EventData {
    // 다음 이벤트 읽기
}
```

#### 2.2 CGO로 Rust 바인딩 호출

```go
// backend/event_ring_reader.go
package main

/*
#cgo LDFLAGS: -L./lib -lmonad_event_ring
#include "monad_event_ring.h"
*/
import "C"

type EventRingReader struct {
    reader *C.EventRingReader
}

func OpenEventRing(path string) (*EventRingReader, error) {
    cPath := C.CString(path)
    defer C.free(unsafe.Pointer(cPath))

    reader := C.monad_event_ring_open(cPath)
    if reader == nil {
        return nil, fmt.Errorf("failed to open event ring")
    }

    return &EventRingReader{reader: reader}, nil
}

func (r *EventRingReader) ReadNext() (*ExecEvent, error) {
    event := C.monad_event_ring_read_next(r.reader)
    return parseExecEvent(event), nil
}
```

#### 2.3 이벤트별 카운터 집계

```go
func (r *EventRingReader) ProcessEvents() {
    for {
        event, err := r.ReadNext()
        if err != nil {
            continue
        }

        switch event.Type {
        case EventTypeTxnHeaderStart:
            waterfallMetrics.NetRPCReceived.Add(1)

        case EventTypeTxnReject:
            rejectReason := event.Data.(TxnReject).Reason
            switch rejectReason {
            case RejectReasonSignature:
                waterfallMetrics.VerifySigFailed.Add(1)
            case RejectReasonNonce:
                waterfallMetrics.VerifyNonceFailed.Add(1)
            case RejectReasonBalance:
                waterfallMetrics.VerifyBalanceFailed.Add(1)
            }

        case EventTypeTxnEvmOutput:
            output := event.Data.(TxnEvmOutput)
            if output.Parallel {
                waterfallMetrics.ExecParallelSuccess.Add(1)
            } else {
                waterfallMetrics.ExecSequentialFallback.Add(1)
            }

        case EventTypeAccountAccess:
            waterfallMetrics.ExecStateReads.Add(1)

        case EventTypeStorageAccess:
            waterfallMetrics.ExecStateWrites.Add(1)

        case EventTypeBlockFinalized:
            waterfallMetrics.BlockFinalized.Add(1)
        }
    }
}
```

---

### **방법 3: JSON-RPC 확장 (간단)**

Monad가 커스텀 RPC 메소드를 제공하는 경우:

```go
// backend/monad_rpc_extended.go

func (c *MonadClient) GetDetailedMetrics() (*DetailedMetrics, error) {
    // 커스텀 RPC 메소드 호출
    resp, err := c.rpcCall(c.ExecutionRPCUrl, "monad_getMetrics", []interface{}{})
    if err != nil {
        return nil, err
    }

    var metrics DetailedMetrics
    if err := json.Unmarshal(resp, &metrics); err != nil {
        return nil, err
    }

    return &metrics, nil
}

type DetailedMetrics struct {
    TxPool struct {
        RPCReceived   int64 `json:"rpc_received"`
        P2PReceived   int64 `json:"p2p_received"`
        SigFailed     int64 `json:"sig_failed"`
        NonceFailed   int64 `json:"nonce_failed"`
    } `json:"txpool"`

    Executor struct {
        ParallelSuccess   int64 `json:"parallel_success"`
        SequentialFallback int64 `json:"sequential_fallback"`
    } `json:"executor"`
}
```

---

## 구현 우선순위

### Phase 1 (가장 쉬움): ⭐
- [ ] Monad Prometheus 엔드포인트 찾기
- [ ] Go Prometheus client로 메트릭 scrape
- [ ] 기존 추정치를 실제 값으로 교체

**예상 작업 시간**: 2-3시간

### Phase 2 (중간):
- [ ] JSON-RPC 커스텀 메소드 확인
- [ ] 추가 메트릭 엔드포인트 연동

**예상 작업 시간**: 4-6시간

### Phase 3 (고급): 🔥
- [ ] Event Ring Rust FFI 바인딩 작성
- [ ] CGO로 Go에서 Event Ring 읽기
- [ ] 이벤트별 실시간 집계

**예상 작업 시간**: 1-2일

---

## 체크리스트: 실제 데이터로 전환

### 준비 사항:
- [ ] Monad 노드가 실행 중인지 확인
- [ ] Prometheus 메트릭 포트 확인 (예: 9090, 6060)
- [ ] 메트릭이 활성화되어 있는지 확인

### 단계별 확인:
```bash
# 1. Monad 프로세스 확인
ps aux | grep monad

# 2. 메트릭 엔드포인트 확인
curl http://localhost:9090/metrics | head -50

# 3. 특정 메트릭 검색
curl http://localhost:9090/metrics | grep -E "txpool|executor|consensus"

# 4. WebSocket 연결 확인
websocat ws://localhost:8081

# 5. Event Ring 소켓 확인
ls -la /home/monad/monad-bft/*.sock
```

### 테스트:
```bash
# 대시보드에서 실제 데이터 확인
./monad-dashboard

# 브라우저에서 확인:
# http://localhost:4000
# TPU Waterfall의 숫자들이 실제 변동하는지 확인
```

---

## 예상 개선 효과

### Before (추정):
```
RPC: 371 (70% 가정)
Sig Failed: 18 (5% 가정)
Parallel: 208 (85% 가정)
```

### After (실제):
```
RPC: 245 (실제 측정값)
Sig Failed: 3 (실제 측정값)
Parallel: 187 (실제 측정값)
```

실제 데이터를 사용하면:
- ✅ 정확한 병목 지점 파악
- ✅ 실제 성능 모니터링
- ✅ 최적화 효과 측정 가능
- ✅ Production 운영에 적합

---

## 문의사항

Monad 팀에 확인이 필요한 사항:
1. Prometheus 메트릭 엔드포인트 주소
2. 제공되는 메트릭 이름 목록
3. Event Ring 접근 방법 및 문서
4. 커스텀 RPC 메소드 목록 (있는 경우)

연락처: Monad Discord / Docs / GitHub
