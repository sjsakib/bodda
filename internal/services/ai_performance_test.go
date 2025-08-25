package services

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

// BenchmarkIterativeProcessor_Creation benchmarks processor creation performance
func BenchmarkIterativeProcessor_Creation(b *testing.B) {
	msgCtx := &MessageContext{
		UserID:    "bench-user",
		SessionID: "bench-session",
		Message:   "Benchmark message",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		processor := NewIterativeProcessor(msgCtx, func(string) {})
		_ = processor // Prevent optimization
	}
}

// BenchmarkIterativeProcessor_ToolResultAccumulation benchmarks tool result accumulation
func BenchmarkIterativeProcessor_ToolResultAccumulation(b *testing.B) {
	processor := NewIterativeProcessor(&MessageContext{}, func(string) {})

	// Pre-create tool results to avoid allocation overhead in benchmark
	toolResults := make([][]ToolResult, b.N)
	for i := 0; i < b.N; i++ {
		toolResults[i] = []ToolResult{
			{
				ToolCallID: fmt.Sprintf("call-%d-1", i),
				Content:    fmt.Sprintf("Result %d-1", i),
			},
			{
				ToolCallID: fmt.Sprintf("call-%d-2", i),
				Content:    fmt.Sprintf("Result %d-2", i),
			},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		processor.AddToolResults(toolResults[i])
	}
}

// BenchmarkAIService_ProgressMessageGeneration benchmarks progress message generation
func BenchmarkAIService_ProgressMessageGeneration(b *testing.B) {
	service := setupTestAIService()
	processor := NewIterativeProcessor(&MessageContext{}, func(string) {})

	toolCalls := []openai.ToolCall{
		{Function: openai.FunctionCall{Name: "get-athlete-profile"}},
		{Function: openai.FunctionCall{Name: "get-recent-activities"}},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		processor.CurrentRound = i % 5
		_ = service.getCoachingProgressMessage(processor, toolCalls)
	}
}

// BenchmarkAIService_AnalysisDepthAssessment benchmarks analysis depth assessment
func BenchmarkAIService_AnalysisDepthAssessment(b *testing.B) {
	service := setupTestAIService()

	// Create processor with realistic tool results
	processor := &IterativeProcessor{
		ToolResults: [][]ToolResult{
			{
				{Content: `{"firstname": "Test", "ftp": 250}`, Error: ""},
				{Content: `[{"distance": 5000, "type": "Run"}]`, Error: ""},
			},
			{
				{Content: `{"description": "Great run", "calories": 350}`, Error: ""},
				{Content: `{"heartrate": [120, 130, 140], "watts": [200, 220, 240]}`, Error: ""},
			},
		},
	}

	currentCalls := []openai.ToolCall{
		{Function: openai.FunctionCall{Name: "get-activity-streams"}},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = service.assessAnalysisDepth(processor, currentCalls)
	}
}

// BenchmarkAIService_ShouldContinueAnalysis benchmarks continue analysis decision
func BenchmarkAIService_ShouldContinueAnalysis(b *testing.B) {
	service := setupTestAIService()

	processor := &IterativeProcessor{
		MaxRounds:    5,
		CurrentRound: 2,
		ToolResults: [][]ToolResult{
			{{Content: `{"firstname": "Test"}`, Error: ""}},
			{{Content: `[{"distance": 5000}]`, Error: ""}},
		},
	}

	toolCalls := []openai.ToolCall{
		{Function: openai.FunctionCall{Name: "get-activity-details"}},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = service.shouldContinueAnalysis(processor, toolCalls, false)
	}
}

// TestAIService_ConcurrentProcessorUsage tests concurrent usage of processors
func TestAIService_ConcurrentProcessorUsage(t *testing.T) {
	t.Run("concurrent processor creation", func(t *testing.T) {
		const numGoroutines = 100
		const processorsPerGoroutine = 10

		var wg sync.WaitGroup
		processors := make([][]*IterativeProcessor, numGoroutines)

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				processors[index] = make([]*IterativeProcessor, processorsPerGoroutine)
				for j := 0; j < processorsPerGoroutine; j++ {
					msgCtx := &MessageContext{
						UserID:    fmt.Sprintf("user-%d-%d", index, j),
						SessionID: fmt.Sprintf("session-%d-%d", index, j),
						Message:   "Concurrent test message",
					}
					processors[index][j] = NewIterativeProcessor(msgCtx, func(string) {})
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		// Should complete quickly even with high concurrency
		assert.Less(t, duration, 100*time.Millisecond, "Concurrent processor creation should be fast")

		// Verify all processors were created correctly
		for i, group := range processors {
			for j, processor := range group {
				assert.NotNil(t, processor, "Processor [%d][%d] should not be nil", i, j)
				assert.Equal(t, 5, processor.MaxRounds)
				assert.Equal(t, 0, processor.CurrentRound)
				assert.Equal(t, fmt.Sprintf("user-%d-%d", i, j), processor.Context.UserID)
			}
		}
	})

	t.Run("concurrent progress message generation", func(t *testing.T) {
		service := setupTestAIService()
		const numGoroutines = 50

		var wg sync.WaitGroup
		messages := make([][]string, numGoroutines)

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				processor := NewIterativeProcessor(&MessageContext{}, func(string) {})
				messages[index] = make([]string, 10)

				toolCalls := []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "get-athlete-profile"}},
				}

				for j := 0; j < 10; j++ {
					processor.CurrentRound = j % 5
					messages[index][j] = service.getCoachingProgressMessage(processor, toolCalls)
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		// Should handle concurrent message generation efficiently
		assert.Less(t, duration, 200*time.Millisecond, "Concurrent message generation should be fast")

		// Verify all messages are valid
		for i, group := range messages {
			for j, message := range group {
				assert.NotEmpty(t, message, "Message [%d][%d] should not be empty", i, j)
				assert.NotContains(t, message, "API", "Message [%d][%d] should not contain API", i, j)
			}
		}
	})

	t.Run("concurrent tool result accumulation", func(t *testing.T) {
		const numGoroutines = 20
		const resultsPerGoroutine = 50

		var wg sync.WaitGroup
		processors := make([]*IterativeProcessor, numGoroutines)

		// Initialize processors
		for i := 0; i < numGoroutines; i++ {
			processors[i] = NewIterativeProcessor(&MessageContext{}, func(string) {})
		}

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				processor := processors[index]
				for j := 0; j < resultsPerGoroutine; j++ {
					results := []ToolResult{
						{
							ToolCallID: fmt.Sprintf("call-%d-%d", index, j),
							Content:    fmt.Sprintf("Result from goroutine %d, iteration %d", index, j),
						},
					}
					processor.AddToolResults(results)
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		// Should handle concurrent accumulation efficiently
		assert.Less(t, duration, 100*time.Millisecond, "Concurrent tool result accumulation should be fast")

		// Verify final states
		for i, processor := range processors {
			assert.Equal(t, resultsPerGoroutine, processor.CurrentRound, "Processor %d should have correct round count", i)
			assert.Equal(t, resultsPerGoroutine, processor.GetTotalToolCalls(), "Processor %d should have correct total calls", i)
			assert.Len(t, processor.ToolResults, resultsPerGoroutine, "Processor %d should have correct results length", i)
		}
	})
}

// TestAIService_MemoryUsage tests memory usage patterns
func TestAIService_MemoryUsage(t *testing.T) {
	t.Run("processor memory usage is reasonable", func(t *testing.T) {
		runtime.GC() // Clean up before measurement
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Create many processors
		processors := make([]*IterativeProcessor, 1000)
		for i := 0; i < 1000; i++ {
			msgCtx := &MessageContext{
				UserID:    fmt.Sprintf("mem-test-%d", i),
				SessionID: fmt.Sprintf("session-%d", i),
				Message:   "Memory test message",
			}
			processors[i] = NewIterativeProcessor(msgCtx, func(string) {})
		}

		runtime.GC() // Force garbage collection
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		memoryUsed := m2.Alloc - m1.Alloc

		// Should use reasonable amount of memory (less than 1MB for 1000 processors)
		assert.Less(t, memoryUsed, uint64(1024*1024), "Memory usage should be reasonable")

		// Verify processors are functional
		for i, processor := range processors {
			assert.NotNil(t, processor, "Processor %d should not be nil", i)
			assert.Equal(t, 0, processor.CurrentRound)
		}
	})

	t.Run("tool result accumulation memory growth", func(t *testing.T) {
		processor := NewIterativeProcessor(&MessageContext{}, func(string) {})

		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Add substantial tool results
		for round := 0; round < 100; round++ {
			results := make([]ToolResult, 5)
			for i := 0; i < 5; i++ {
				results[i] = ToolResult{
					ToolCallID: fmt.Sprintf("call-%d-%d", round, i),
					Content:    fmt.Sprintf("Large content for round %d, tool %d with additional data to simulate realistic payload size", round, i),
				}
			}
			processor.AddToolResults(results)
		}

		runtime.GC()
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		memoryUsed := m2.Alloc - m1.Alloc

		// Should use reasonable memory even with large datasets (less than 5MB)
		assert.Less(t, memoryUsed, uint64(5*1024*1024), "Memory growth should be reasonable")

		// Verify data integrity
		assert.Equal(t, 100, processor.CurrentRound)
		assert.Equal(t, 500, processor.GetTotalToolCalls())
	})
}

// TestAIService_ScalabilityLimits tests scalability limits
func TestAIService_ScalabilityLimits(t *testing.T) {
	t.Run("handles large number of tool results efficiently", func(t *testing.T) {
		processor := NewIterativeProcessor(&MessageContext{}, func(string) {})

		start := time.Now()

		// Simulate very large analysis session
		for round := 0; round < 1000; round++ {
			results := []ToolResult{
				{
					ToolCallID: fmt.Sprintf("call-%d", round),
					Content:    fmt.Sprintf("Result %d", round),
				},
			}
			processor.AddToolResults(results)
		}

		duration := time.Since(start)

		// Should handle large datasets efficiently (under 100ms for 1000 rounds)
		assert.Less(t, duration, 100*time.Millisecond, "Large dataset handling should be efficient")

		// Verify final state
		assert.Equal(t, 1000, processor.CurrentRound)
		assert.Equal(t, 1000, processor.GetTotalToolCalls())
	})

	t.Run("analysis depth assessment scales with data size", func(t *testing.T) {
		service := setupTestAIService()

		// Create processor with very large tool result history
		processor := &IterativeProcessor{
			ToolResults: make([][]ToolResult, 500),
		}

		// Fill with realistic but large dataset
		for round := 0; round < 500; round++ {
			processor.ToolResults[round] = []ToolResult{
				{Content: fmt.Sprintf(`{"firstname": "User%d", "ftp": %d}`, round, 200+round), Error: ""},
				{Content: fmt.Sprintf(`[{"distance": %d, "type": "Run"}]`, 5000+round*10), Error: ""},
			}
		}

		currentCalls := []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-activity-streams"}},
		}

		start := time.Now()

		// Should still be fast even with large history
		depth := service.assessAnalysisDepth(processor, currentCalls)

		duration := time.Since(start)

		// Should remain fast even with large datasets (under 10ms)
		assert.Less(t, duration, 10*time.Millisecond, "Analysis depth assessment should scale well")
		assert.Equal(t, 3, depth) // Profile + activities + streams (current)
	})

	t.Run("progress message generation performance with high frequency", func(t *testing.T) {
		service := setupTestAIService()
		processor := NewIterativeProcessor(&MessageContext{}, func(string) {})

		toolCalls := []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-recent-activities"}},
		}

		start := time.Now()

		// Generate many messages rapidly
		messages := make([]string, 10000)
		for i := 0; i < 10000; i++ {
			processor.CurrentRound = i % 5
			messages[i] = service.getCoachingProgressMessage(processor, toolCalls)
		}

		duration := time.Since(start)

		// Should handle high frequency generation (under 100ms for 10k messages)
		assert.Less(t, duration, 100*time.Millisecond, "High frequency message generation should be fast")

		// Verify message quality
		for i, message := range messages {
			if i%1000 == 0 { // Sample check to avoid slowing down test
				assert.NotEmpty(t, message)
				assert.NotContains(t, message, "API")
			}
		}
	})
}

// TestAIService_ResourceCleanup tests resource cleanup and garbage collection
func TestAIService_ResourceCleanup(t *testing.T) {
	t.Run("processors can be garbage collected", func(t *testing.T) {
		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Create and discard many processors
		func() {
			processors := make([]*IterativeProcessor, 1000)
			for i := 0; i < 1000; i++ {
				msgCtx := &MessageContext{
					UserID:    fmt.Sprintf("gc-test-%d", i),
					SessionID: fmt.Sprintf("session-%d", i),
					Message:   "GC test message",
				}
				processors[i] = NewIterativeProcessor(msgCtx, func(string) {})

				// Add some data to make them more substantial
				results := []ToolResult{
					{ToolCallID: fmt.Sprintf("call-%d", i), Content: fmt.Sprintf("Content %d", i)},
				}
				processors[i].AddToolResults(results)
			}
			// processors go out of scope here
		}()

		// Force garbage collection
		runtime.GC()
		runtime.GC() // Call twice to ensure cleanup

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		// Memory should be cleaned up (allow some variance for GC behavior)
		memoryDiff := int64(m2.Alloc) - int64(m1.Alloc)
		assert.Less(t, memoryDiff, int64(500*1024), "Memory should be cleaned up after GC")
	})

	t.Run("large tool results can be garbage collected", func(t *testing.T) {
		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Create processor with large data that goes out of scope
		func() {
			processor := NewIterativeProcessor(&MessageContext{}, func(string) {})

			// Add large tool results
			for round := 0; round < 100; round++ {
				results := make([]ToolResult, 10)
				for i := 0; i < 10; i++ {
					// Create large content strings
					largeContent := fmt.Sprintf("Large content for round %d, tool %d: %s", round, i, 
						string(make([]byte, 1024))) // 1KB per result
					results[i] = ToolResult{
						ToolCallID: fmt.Sprintf("call-%d-%d", round, i),
						Content:    largeContent,
					}
				}
				processor.AddToolResults(results)
			}
			// processor goes out of scope here
		}()

		// Force garbage collection
		runtime.GC()
		runtime.GC()

		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		// Large data should be cleaned up
		memoryDiff := int64(m2.Alloc) - int64(m1.Alloc)
		assert.Less(t, memoryDiff, int64(5*1024*1024), "Large tool results should be garbage collected")
	})
}

// TestAIService_EdgeCasePerformance tests performance under edge cases
func TestAIService_EdgeCasePerformance(t *testing.T) {
	t.Run("empty tool results performance", func(t *testing.T) {
		service := setupTestAIService()
		processor := &IterativeProcessor{
			ToolResults: [][]ToolResult{},
		}

		start := time.Now()

		// Should handle empty data efficiently
		for i := 0; i < 1000; i++ {
			_ = service.assessAnalysisDepth(processor, []openai.ToolCall{})
		}

		duration := time.Since(start)
		assert.Less(t, duration, 10*time.Millisecond, "Empty data handling should be very fast")
	})

	t.Run("maximum rounds edge case", func(t *testing.T) {
		processor := &IterativeProcessor{
			MaxRounds:    1000, // Very high limit
			CurrentRound: 999,  // Near limit
		}

		start := time.Now()

		// Should handle edge cases efficiently
		for i := 0; i < 1000; i++ {
			_ = processor.ShouldContinue(true)
		}

		duration := time.Since(start)
		assert.Less(t, duration, 5*time.Millisecond, "Edge case handling should be fast")
	})

	t.Run("very long content strings", func(t *testing.T) {
		service := setupTestAIService()

		// Create very long content strings
		longContent := string(make([]byte, 100*1024)) // 100KB content
		processor := &IterativeProcessor{
			ToolResults: [][]ToolResult{
				{
					{Content: longContent, Error: ""},
				},
			},
		}

		start := time.Now()

		// Should handle long content efficiently
		for i := 0; i < 100; i++ {
			_ = service.assessAnalysisDepth(processor, []openai.ToolCall{})
		}

		duration := time.Since(start)
		assert.Less(t, duration, 50*time.Millisecond, "Long content handling should be reasonably fast")
	})
}