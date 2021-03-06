/*
Copyright 2018 Google Inc. All Rights Reserved.
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

package autoscaler

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/golang/glog"
	"github.com/josephburnett/k8sflag/pkg/k8sflag"
	"github.com/knative/serving/pkg/logging"
)

var (
	testLogger = zap.NewNop().Sugar()
	testCtx    = logging.WithLogger(context.TODO(), testLogger)
)

func TestAutoscaler_NoData_NoAutoscale(t *testing.T) {
	a := newTestAutoscaler(10.0)
	a.expectScale(t, time.Now(), 0, false)
}

func TestAutoscaler_StableMode_NoChange(t *testing.T) {
	a := newTestAutoscaler(10.0)
	now := a.recordLinearSeries(
		time.Now(),
		linearSeries{
			startConcurrency: 10,
			endConcurrency:   10,
			durationSeconds:  60,
			podCount:         10,
		})
	a.expectScale(t, now, 10, true)
}

func TestAutoscaler_StableMode_SlowIncrease(t *testing.T) {
	a := newTestAutoscaler(10.0)
	now := a.recordLinearSeries(
		time.Now(),
		linearSeries{
			startConcurrency: 10,
			endConcurrency:   20,
			durationSeconds:  60,
			podCount:         10,
		})
	a.expectScale(t, now, 15, true)
}

func TestAutoscaler_StableMode_SlowDecrease(t *testing.T) {
	a := newTestAutoscaler(10.0)
	now := a.recordLinearSeries(
		time.Now(),
		linearSeries{
			startConcurrency: 20,
			endConcurrency:   10,
			durationSeconds:  60,
			podCount:         10,
		})
	a.expectScale(t, now, 15, true)
}

func TestAutoscaler_StableModeLowPodCount_NoChange(t *testing.T) {
	a := newTestAutoscaler(10.0)
	now := a.recordLinearSeries(
		time.Now(),
		linearSeries{
			startConcurrency: 10,
			endConcurrency:   10,
			durationSeconds:  60,
			podCount:         1,
		})
	a.expectScale(t, now, 1, true)
}

func TestAutoscaler_StableModeNoTraffic_ScaleToOne(t *testing.T) {
	a := newTestAutoscaler(10.0)
	now := a.recordLinearSeries(
		time.Now(),
		linearSeries{
			startConcurrency: 0,
			endConcurrency:   0,
			durationSeconds:  60,
			podCount:         2,
		})
	a.expectScale(t, now, 1, true)
}

func TestAutoscaler_StableModeNoTraffic_ScaleToZero(t *testing.T) {
	a := newTestAutoscaler(10.0)
	now := a.recordLinearSeries(
		time.Now(),
		linearSeries{
			startConcurrency: 1,
			endConcurrency:   1,
			durationSeconds:  60,
			podCount:         1,
		})

	a.expectScale(t, now, 1, true)
	now = a.recordLinearSeries(
		now,
		linearSeries{
			startConcurrency: 0,
			endConcurrency:   0,
			durationSeconds:  300, // 5 minutes
			podCount:         1,
		})
	a.expectScale(t, now, 0, true)

}
func TestAutoscaler_PanicMode_DoublePodCount(t *testing.T) {
	a := newTestAutoscaler(10.0)
	now := a.recordLinearSeries(
		time.Now(),
		linearSeries{
			startConcurrency: 10,
			endConcurrency:   10,
			durationSeconds:  60,
			podCount:         10,
		})
	now = a.recordLinearSeries(
		now,
		linearSeries{
			startConcurrency: 20,
			endConcurrency:   20,
			durationSeconds:  6,
			podCount:         10,
		})
	a.expectScale(t, now, 20, true)
}

// QPS is increasing exponentially. Each scaling event bring concurrency
// back to the target level (1.0) but then traffic continues to increase.
// At 1296 QPS traffic stablizes.
func TestAutoscaler_PanicModeExponential_TrackAndStablize(t *testing.T) {
	a := newTestAutoscaler(1.0)
	now := a.recordLinearSeries(
		time.Now(),
		linearSeries{
			startConcurrency: 1,
			endConcurrency:   10,
			durationSeconds:  6,
			podCount:         1,
		})
	a.expectScale(t, now, 6, true)
	now = a.recordLinearSeries(
		now,
		linearSeries{
			startConcurrency: 1,
			endConcurrency:   10,
			durationSeconds:  6,
			podCount:         6,
		})
	a.expectScale(t, now, 36, true)
	now = a.recordLinearSeries(
		now,
		linearSeries{
			startConcurrency: 1,
			endConcurrency:   10,
			durationSeconds:  6,
			podCount:         36,
		})
	a.expectScale(t, now, 216, true)
	now = a.recordLinearSeries(
		now,
		linearSeries{
			startConcurrency: 1,
			endConcurrency:   10,
			durationSeconds:  6,
			podCount:         216,
		})
	a.expectScale(t, now, 1296, true)
	now = a.recordLinearSeries(
		now,
		linearSeries{
			startConcurrency: 1,
			endConcurrency:   1, // achieved desired concurrency
			durationSeconds:  6,
			podCount:         1296,
		})
	a.expectScale(t, now, 1296, true)
}

func TestAutoscaler_PanicThenUnPanic_ScaleDown(t *testing.T) {
	a := newTestAutoscaler(10.0)
	now := a.recordLinearSeries(
		time.Now(),
		linearSeries{
			startConcurrency: 10,
			endConcurrency:   10,
			durationSeconds:  60,
			podCount:         10,
		})
	a.expectScale(t, now, 10, true)
	now = a.recordLinearSeries(
		now,
		linearSeries{
			startConcurrency: 100,
			endConcurrency:   100,
			durationSeconds:  6,
			podCount:         10,
		})
	a.expectScale(t, now, 100, true)
	now = a.recordLinearSeries(
		now,
		linearSeries{
			startConcurrency: 1, // traffic drops off
			endConcurrency:   1,
			durationSeconds:  30,
			podCount:         100,
		})
	a.expectScale(t, now, 100, true) // still in panic mode--no decrease
	now = a.recordLinearSeries(
		now,
		linearSeries{
			startConcurrency: 1,
			endConcurrency:   1,
			durationSeconds:  31,
			podCount:         100,
		})
	a.expectScale(t, now, 10, true) // back to stable mode
}

// Autoscaler should drop data after 60 seconds.
func TestAutoscaler_Stats_TrimAfterStableWindow(t *testing.T) {
	a := newTestAutoscaler(10.0)
	now := a.recordLinearSeries(
		time.Now(),
		linearSeries{
			startConcurrency: 10,
			endConcurrency:   10,
			durationSeconds:  60,
			podCount:         1,
		})
	a.expectScale(t, now, 1, true)
	if len(a.stats) != 60 {
		t.Errorf("Unexpected stat count. Expected 60. Got %v.", len(a.stats))
	}
	now = now.Add(time.Minute)
	a.expectScale(t, now, 0, false)
	if len(a.stats) != 0 {
		t.Errorf("Unexpected stat count. Expected 0. Got %v.", len(a.stats))
	}
}

type linearSeries struct {
	startConcurrency int
	endConcurrency   int
	durationSeconds  int
	podCount         int
}

type mockReporter struct{}

func (r *mockReporter) Report(m Measurement, v int64) error {
	return nil
}

func newTestAutoscaler(targetConcurrency float64) *Autoscaler {
	stableWindow := 60 * time.Second
	panicWindow := 6 * time.Second
	scaleToZeroThreshold := 5 * time.Minute
	config := Config{
		TargetConcurrency:    k8sflag.Float64("target-concurrency", targetConcurrency),
		MaxScaleUpRate:       k8sflag.Float64("max-scale-up-rate", 10.0),
		StableWindow:         k8sflag.Duration("stable-window", &stableWindow),
		PanicWindow:          k8sflag.Duration("panic-window", &panicWindow),
		ScaleToZeroThreshold: k8sflag.Duration("scale-to-zero-threshold", &scaleToZeroThreshold),
	}
	return NewAutoscaler(config, &mockReporter{})
}

// Record a data point every second, for every pod, for duration of the
// linear series, on the line from start to end concurrency.
func (a *Autoscaler) recordLinearSeries(now time.Time, s linearSeries) time.Time {
	points := make([]int32, 0)
	for i := 1; i <= s.durationSeconds; i++ {
		points = append(points, int32(float64(s.startConcurrency)+float64(s.endConcurrency-s.startConcurrency)*(float64(i)/float64(s.durationSeconds))))
	}
	glog.Infof("Recording points: %v.", points)
	for _, point := range points {
		t := now
		now = now.Add(time.Second)
		for j := 1; j <= s.podCount; j++ {
			t = t.Add(time.Millisecond)
			stat := Stat{
				Time:                      &t,
				PodName:                   fmt.Sprintf("pod-%v", j),
				AverageConcurrentRequests: float64(point),
			}
			a.Record(testCtx, stat)
		}
	}
	return now
}

func (a *Autoscaler) expectScale(t *testing.T, now time.Time, expectScale int32, expectOk bool) {
	scale, ok := a.Scale(testCtx, now)
	if ok != expectOk {
		t.Errorf("Unexpected autoscale decison. Expected %v. Got %v.", expectOk, ok)
	}
	if scale != expectScale {
		t.Errorf("Unexpected scale. Expected %v. Got %v.", expectScale, scale)
	}
}
