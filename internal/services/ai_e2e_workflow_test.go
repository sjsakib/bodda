package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"bodda/internal/models"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestAIService_EndToEndCoachingWorkflows tests complete realistic coaching scenarios
func TestAIService_EndToEndCoachingWorkflows(t *testing.T) {
	t.Run("complete marathon training analysis", func(t *testing.T) {
		service := setupTestAIService()
		ctx := context.Background()

		// Setup marathon training scenario
		user := &models.User{
			ID:          "marathon-runner",
			AccessToken: "marathon-token",
		}

		existingLogbook := &models.AthleteLogbook{
			UserID: "marathon-runner",
			Content: `{
				"personal_info": {
					"name": "Alex Marathon",
					"age": 32,
					"experience": "intermediate",
					"goal_race": "Boston Marathon in 16 weeks"
				},
				"training_history": {
					"weekly_volume": "60-70km",
					"longest_run": "32km",
					"recent_races": ["Half Marathon: 1:25:30"]
				},
				"goals": {
					"marathon_time": "sub-3:00",
					"weekly_peak": "80km",
					"key_workouts": ["tempo runs", "long runs", "track intervals"]
				}
			}`,
		}

		msgCtx := &MessageContext{
			UserID:              "marathon-runner",
			SessionID:           "marathon-analysis",
			Message:             "I'm 8 weeks into my marathon training. How am I progressing and what should I focus on?",
			ConversationHistory: []*models.Message{},
			AthleteLogbook:      existingLogbook,
			User:                user,
		}

		// Mock comprehensive marathon training data
		marathonProfile := &StravaAthlete{
			ID:        444444,
			Firstname: "Alex",
			Lastname:  "Marathon",
			Sex:       "M",
			Weight:    72.0,
			FTP:       280,
		}
		service.mockStrava.On("GetAthleteProfile", "marathon-token").Return(marathonProfile, nil)

		marathonActivities := []*StravaActivity{
			{
				ID:           4001,
				Name:         "Long Run - 28km",
				Distance:     28000.0,
				MovingTime:   6720, // 1:52:00 (4:00/km pace)
				Type:         "Run",
				AverageSpeed: 4.17,
				StartDate:    "2024-01-20T07:00:00Z",
			},
			{
				ID:           4002,
				Name:         "Tempo Run - 12km",
				Distance:     12000.0,
				MovingTime:   2700, // 45:00 (3:45/km pace)
				Type:         "Run",
				AverageSpeed: 4.44,
				StartDate:    "2024-01-18T18:00:00Z",
			},
			{
				ID:           4003,
				Name:         "Track Intervals - 8x800m",
				Distance:     10000.0,
				MovingTime:   2400, // 40:00 including recovery
				Type:         "Run",
				AverageSpeed: 4.17,
				StartDate:    "2024-01-16T17:30:00Z",
			},
			{
				ID:           4004,
				Name:         "Easy Recovery Run",
				Distance:     8000.0,
				MovingTime:   2400, // 40:00 (5:00/km pace)
				Type:         "Run",
				AverageSpeed: 3.33,
				StartDate:    "2024-01-15T07:00:00Z",
			},
		}
		service.mockStrava.On("GetActivities", "marathon-token", mock.AnythingOfType("ActivityParams")).Return(marathonActivities, nil)

		// Mock detailed analysis for key workouts
		longRunDetail := &StravaActivityDetail{
			StravaActivity: *marathonActivities[0],
			Description:    "Felt strong throughout. Negative split. Good marathon pace practice.",
			Calories:       1680.0,
		}
		service.mockStrava.On("GetActivityDetail", "marathon-token", int64(4001)).Return(longRunDetail, nil)

		tempoRunDetail := &StravaActivityDetail{
			StravaActivity: *marathonActivities[1],
			Description:    "Threshold pace workout. Maintained effort well.",
			Calories:       720.0,
		}
		service.mockStrava.On("GetActivityDetail", "marathon-token", int64(4002)).Return(tempoRunDetail, nil)

		intervalDetail := &StravaActivityDetail{
			StravaActivity: *marathonActivities[2],
			Description:    "8x800m @ 3:10 pace with 90s recovery. Hit all splits.",
			Calories:       600.0,
		}
		service.mockStrava.On("GetActivityDetail", "marathon-token", int64(4003)).Return(intervalDetail, nil)

		// Mock streams for key workout analysis
		longRunStreams := &StravaStreams{
			Time:      []int{0, 1680, 3360, 5040, 6720}, // Every 28 minutes
			Distance:  []float64{0, 7000, 14000, 21000, 28000},
			Heartrate: []int{145, 155, 160, 165, 162}, // Gradual increase, slight drop at end
			Watts:     []int{260, 270, 275, 280, 275},
		}
		service.mockStrava.On("GetActivityStreams", "marathon-token", int64(4001), mock.AnythingOfType("[]string"), mock.AnythingOfType("string")).Return(longRunStreams, nil)

		intervalStreams := &StravaStreams{
			Time:      []int{0, 190, 380, 570, 760, 950, 1140, 1330, 1520, 2400}, // Interval + recovery pattern
			Distance:  []float64{0, 800, 1200, 2000, 2400, 3200, 3600, 4400, 4800, 10000},
			Heartrate: []int{120, 175, 140, 180, 145, 182, 148, 185, 150, 130}, // Interval spikes
			Watts:     []int{200, 350, 220, 360, 230, 365, 235, 370, 240, 200},
		}
		service.mockStrava.On("GetActivityStreams", "marathon-token", int64(4003), mock.AnythingOfType("[]string"), mock.AnythingOfType("string")).Return(intervalStreams, nil)

		// Mock comprehensive logbook update
		updatedMarathonLogbook := &models.AthleteLogbook{
			UserID: "marathon-runner",
			Content: `{
				"personal_info": {
					"name": "Alex Marathon",
					"age": 32,
					"experience": "intermediate",
					"goal_race": "Boston Marathon in 8 weeks"
				},
				"training_analysis": {
					"current_phase": "peak training",
					"weekly_volume": "68km average last 4 weeks",
					"longest_run": "28km - excellent progression",
					"tempo_fitness": "strong - maintaining 3:45/km for 12km",
					"speed_work": "on target - 800m intervals at 3:10 pace"
				},
				"performance_indicators": {
					"aerobic_fitness": "excellent - HR control during long runs",
					"lactate_threshold": "strong - tempo runs well executed",
					"neuromuscular_power": "good - interval splits consistent",
					"recovery": "adequate - easy runs at appropriate pace"
				},
				"recommendations": {
					"next_2_weeks": "maintain current volume, add one more 30km+ run",
					"taper_strategy": "reduce volume by 25% starting week 6",
					"race_strategy": "target 4:15/km for first half, negative split if feeling strong",
					"areas_to_watch": "monitor recovery between hard sessions"
				}
			}`,
		}
		service.mockLogbook.On("UpdateLogbook", ctx, "marathon-runner", mock.AnythingOfType("string")).Return(updatedMarathonLogbook, nil)

		// Execute comprehensive multi-round analysis
		processor := NewIterativeProcessor(msgCtx, func(string) {})

		// Simulate complete analysis workflow
		analysisRounds := []struct {
			description string
			toolCalls   []openai.ToolCall
			verify      func(t *testing.T, processor *IterativeProcessor)
		}{
			{
				description: "Initial profile assessment",
				toolCalls: []openai.ToolCall{
					{ID: "call-1", Function: openai.FunctionCall{Name: "get-athlete-profile"}},
				},
				verify: func(t *testing.T, p *IterativeProcessor) {
					shouldContinue, reason := service.shouldContinueAnalysis(p, []openai.ToolCall{
						{Function: openai.FunctionCall{Name: "get-recent-activities"}},
					}, false)
					assert.True(t, shouldContinue)
					assert.Equal(t, "continue_analysis", reason)
				},
			},
			{
				description: "Recent training review",
				toolCalls: []openai.ToolCall{
					{ID: "call-2", Function: openai.FunctionCall{Name: "get-recent-activities", Arguments: `{"per_page": 20}`}},
				},
				verify: func(t *testing.T, p *IterativeProcessor) {
					depth := service.assessAnalysisDepth(p, []openai.ToolCall{})
					assert.Equal(t, 2, depth) // Profile + activities
				},
			},
			{
				description: "Key workout analysis",
				toolCalls: []openai.ToolCall{
					{ID: "call-3", Function: openai.FunctionCall{Name: "get-activity-details", Arguments: `{"activity_id": 4001}`}},
					{ID: "call-4", Function: openai.FunctionCall{Name: "get-activity-details", Arguments: `{"activity_id": 4002}`}},
					{ID: "call-5", Function: openai.FunctionCall{Name: "get-activity-details", Arguments: `{"activity_id": 4003}`}},
				},
				verify: func(t *testing.T, p *IterativeProcessor) {
					assert.Equal(t, 3, p.CurrentRound)
					assert.Equal(t, 5, p.GetTotalToolCalls()) // 1 + 1 + 3
				},
			},
			{
				description: "Performance data deep dive",
				toolCalls: []openai.ToolCall{
					{ID: "call-6", Function: openai.FunctionCall{Name: "get-activity-streams", Arguments: `{"activity_id": 4001, "stream_types": ["heartrate", "watts"], "resolution": "high"}`}},
					{ID: "call-7", Function: openai.FunctionCall{Name: "get-activity-streams", Arguments: `{"activity_id": 4003, "stream_types": ["heartrate", "watts"], "resolution": "high"}`}},
				},
				verify: func(t *testing.T, p *IterativeProcessor) {
					shouldContinue := service.toolCallsSuggestDeeperAnalysis([]openai.ToolCall{
						{Function: openai.FunctionCall{Name: "get-activity-streams"}},
					})
					assert.True(t, shouldContinue)
				},
			},
			{
				description: "Comprehensive logbook update",
				toolCalls: []openai.ToolCall{
					{ID: "call-8", Function: openai.FunctionCall{Name: "update-athlete-logbook", Arguments: `{"content": "Comprehensive marathon training analysis..."}`}},
				},
				verify: func(t *testing.T, p *IterativeProcessor) {
					assert.Equal(t, 5, p.CurrentRound)
					assert.Equal(t, 8, p.GetTotalToolCalls())
				},
			},
		}

		// Execute each analysis round
		for i, round := range analysisRounds {
			t.Logf("Executing analysis round %d: %s", i+1, round.description)

			// Verify progress messaging is appropriate
			progressMsg := service.getCoachingProgressMessage(processor, round.toolCalls)
			assert.NotEmpty(t, progressMsg)
			assert.NotContains(t, progressMsg, "API")
			assert.NotContains(t, progressMsg, "executing")
			assert.NotContains(t, progressMsg, "tool")

			// Execute tools for this round
			results, err := service.executeTools(ctx, msgCtx, round.toolCalls)
			assert.NoError(t, err)
			assert.Len(t, results, len(round.toolCalls))

			// Add results and verify state
			processor.AddToolResults(results)
			round.verify(t, processor)
		}

		// Verify final analysis state
		assert.Equal(t, 5, processor.CurrentRound)
		assert.Equal(t, 8, processor.GetTotalToolCalls())

		// Verify comprehensive analysis was performed
		finalDepth := service.assessAnalysisDepth(processor, []openai.ToolCall{})
		assert.Equal(t, 4, finalDepth) // All analysis types completed

		// Verify final decision making
		shouldContinue, reason := service.shouldContinueAnalysis(processor, []openai.ToolCall{}, false)
		assert.False(t, shouldContinue)
		assert.Equal(t, "no_tools", reason)

		// Generate final coaching response
		finalResponse := service.generateFinalResponse(processor, reason, false)
		assert.NotEmpty(t, finalResponse)
		assert.NotContains(t, finalResponse, "API")
		assert.Contains(t, finalResponse, "training") // Should be coaching-focused

		service.mockStrava.AssertExpectations(t)
		service.mockLogbook.AssertExpectations(t)
	})

	t.Run("injury prevention analysis workflow", func(t *testing.T) {
		service := setupTestAIService()
		ctx := context.Background()

		// Setup injury prevention scenario
		user := &models.User{
			ID:          "injury-prevention",
			AccessToken: "prevention-token",
		}

		concernLogbook := &models.AthleteLogbook{
			UserID: "injury-prevention",
			Content: `{
				"personal_info": {
					"name": "Jordan Careful",
					"age": 28,
					"injury_history": ["IT band syndrome 6 months ago", "minor calf strain last year"]
				},
				"current_concerns": {
					"knee_discomfort": "mild soreness after long runs",
					"training_load": "increased volume by 30% in last month",
					"recovery_quality": "sleep has been inconsistent"
				},
				"prevention_goals": {
					"maintain_health": "avoid injury while building fitness",
					"smart_progression": "increase volume safely",
					"monitoring": "watch for early warning signs"
				}
			}`,
		}

		msgCtx := &MessageContext{
			UserID:              "injury-prevention",
			SessionID:           "prevention-analysis",
			Message:             "I've been feeling some knee discomfort after my long runs. Can you analyze my training and help me prevent injury?",
			ConversationHistory: []*models.Message{},
			AthleteLogbook:      concernLogbook,
			User:                user,
		}

		// Mock injury prevention focused data
		preventionProfile := &StravaAthlete{
			ID:        555555,
			Firstname: "Jordan",
			Lastname:  "Careful",
			Sex:       "F",
			Weight:    58.0,
		}
		service.mockStrava.On("GetAthleteProfile", "prevention-token").Return(preventionProfile, nil)

		concerningActivities := []*StravaActivity{
			{
				ID:           5001,
				Name:         "Long Run - Felt knee discomfort",
				Distance:     18000.0,
				MovingTime:   4500, // 1:15:00 (4:10/km)
				Type:         "Run",
				AverageSpeed: 4.0,
				StartDate:    "2024-01-21T08:00:00Z",
			},
			{
				ID:           5002,
				Name:         "Medium Run - Pushed pace",
				Distance:     12000.0,
				MovingTime:   2700, // 45:00 (3:45/km)
				Type:         "Run",
				AverageSpeed: 4.44,
				StartDate:    "2024-01-19T18:00:00Z",
			},
			{
				ID:           5003,
				Name:         "Back-to-back long run",
				Distance:     15000.0,
				MovingTime:   3900, // 1:05:00 (4:20/km)
				Type:         "Run",
				AverageSpeed: 3.85,
				StartDate:    "2024-01-18T08:00:00Z",
			},
		}
		service.mockStrava.On("GetActivities", "prevention-token", mock.AnythingOfType("ActivityParams")).Return(concerningActivities, nil)

		// Mock detailed analysis focusing on load and recovery
		longRunDetail := &StravaActivityDetail{
			StravaActivity: *concerningActivities[0],
			Description:    "Knee started bothering me around 15km mark. Finished but concerned.",
			Calories:       900.0,
		}
		service.mockStrava.On("GetActivityDetail", "prevention-token", int64(5001)).Return(longRunDetail, nil)

		mediumRunDetail := &StravaActivityDetail{
			StravaActivity: *concerningActivities[1],
			Description:    "Felt good during run but legs were tired from yesterday.",
			Calories:       600.0,
		}
		service.mockStrava.On("GetActivityDetail", "prevention-token", int64(5002)).Return(mediumRunDetail, nil)

		// Mock streams showing potential overload patterns
		concerningStreams := &StravaStreams{
			Time:      []int{0, 1125, 2250, 3375, 4500}, // Every 18:45
			Distance:  []float64{0, 4500, 9000, 13500, 18000},
			Heartrate: []int{150, 160, 165, 170, 175}, // Gradual increase - potential fatigue
			Watts:     []int{240, 250, 255, 260, 265}, // Increasing effort for same pace
		}
		service.mockStrava.On("GetActivityStreams", "prevention-token", int64(5001), mock.AnythingOfType("[]string"), mock.AnythingOfType("string")).Return(concerningStreams, nil)

		// Mock prevention-focused logbook update
		preventionLogbook := &models.AthleteLogbook{
			UserID: "injury-prevention",
			Content: `{
				"personal_info": {
					"name": "Jordan Careful",
					"age": 28,
					"injury_history": ["IT band syndrome 6 months ago", "minor calf strain last year"]
				},
				"risk_assessment": {
					"current_risk": "moderate - showing early warning signs",
					"load_analysis": "30% volume increase in 4 weeks - too aggressive",
					"recovery_indicators": "HR drift in long runs suggests fatigue accumulation",
					"biomechanical_stress": "knee discomfort pattern consistent with overuse"
				},
				"prevention_strategy": {
					"immediate_actions": ["reduce weekly volume by 20%", "add extra rest day", "focus on easy pace runs"],
					"monitoring": ["track knee discomfort daily", "monitor HR drift in long runs", "assess sleep quality"],
					"strengthening": ["increase glute and hip stability work", "calf strengthening", "core stability"],
					"recovery_enhancement": ["prioritize sleep consistency", "add gentle stretching routine", "consider massage"]
				},
				"red_flags": {
					"stop_signs": ["sharp pain during running", "pain that worsens during activity", "limping after runs"],
					"caution_signs": ["persistent soreness >24hrs", "compensatory movement patterns", "declining performance"]
				}
			}`,
		}
		service.mockLogbook.On("UpdateLogbook", ctx, "injury-prevention", mock.AnythingOfType("string")).Return(preventionLogbook, nil)

		// Execute prevention-focused analysis
		processor := NewIterativeProcessor(msgCtx, func(string) {})

		// Verify analysis focuses on injury prevention
		preventionRounds := []struct {
			toolCalls []openai.ToolCall
			focus     string
		}{
			{
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "get-athlete-profile"}},
				},
				focus: "baseline assessment",
			},
			{
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "get-recent-activities"}},
				},
				focus: "training load analysis",
			},
			{
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "get-activity-details", Arguments: `{"activity_id": 5001}`}},
					{Function: openai.FunctionCall{Name: "get-activity-details", Arguments: `{"activity_id": 5002}`}},
				},
				focus: "symptom correlation",
			},
			{
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "get-activity-streams", Arguments: `{"activity_id": 5001}`}},
				},
				focus: "physiological stress analysis",
			},
			{
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "update-athlete-logbook"}},
				},
				focus: "prevention strategy documentation",
			},
		}

		for i, round := range preventionRounds {
			t.Logf("Prevention analysis round %d: %s", i+1, round.focus)

			// Verify progress messages are appropriate for injury prevention context
			progressMsg := service.getCoachingProgressMessage(processor, round.toolCalls)
			assert.NotEmpty(t, progressMsg)
			assert.NotContains(t, progressMsg, "API")
			assert.NotContains(t, progressMsg, "performance") // Should focus on health, not performance

			// Execute analysis round
			results, err := service.executeTools(ctx, msgCtx, round.toolCalls)
			assert.NoError(t, err)
			processor.AddToolResults(results)
		}

		// Verify prevention-focused final response
		finalResponse := service.generateFinalResponse(processor, "sufficient_data", false)
		assert.NotEmpty(t, finalResponse)
		assert.NotContains(t, finalResponse, "API")
		// Should contain health/safety focused language
		healthWords := []string{"health", "safe", "prevent", "careful", "recovery", "rest"}
		hasHealthFocus := false
		for _, word := range healthWords {
			if strings.Contains(strings.ToLower(finalResponse), word) {
				hasHealthFocus = true
				break
			}
		}
		assert.True(t, hasHealthFocus, "Response should be health-focused for injury prevention")

		service.mockStrava.AssertExpectations(t)
		service.mockLogbook.AssertExpectations(t)
	})
}

// TestAIService_ErrorRecoveryWorkflows tests error recovery in realistic scenarios
func TestAIService_ErrorRecoveryWorkflows(t *testing.T) {
	t.Run("strava api intermittent failures during analysis", func(t *testing.T) {
		service := setupTestAIService()
		ctx := context.Background()

		user := &models.User{
			ID:          "error-recovery",
			AccessToken: "error-token",
		}

		msgCtx := &MessageContext{
			UserID:    "error-recovery",
			SessionID: "error-session",
			Message:   "Analyze my training progress",
			User:      user,
		}

		// Mock mixed success/failure scenario
		expectedProfile := &StravaAthlete{
			ID:        666666,
			Firstname: "Test",
			Lastname:  "Recovery",
		}
		service.mockStrava.On("GetAthleteProfile", "error-token").Return(expectedProfile, nil).Once()

		// First activities call fails
		service.mockStrava.On("GetActivities", "error-token", mock.AnythingOfType("ActivityParams")).Return(([]*StravaActivity)(nil), fmt.Errorf("Strava API temporarily unavailable")).Once()

		// Second activities call succeeds (simulating recovery)
		expectedActivities := []*StravaActivity{
			{ID: 6001, Name: "Recovery Run", Type: "Run"},
		}
		service.mockStrava.On("GetActivities", "error-token", mock.AnythingOfType("ActivityParams")).Return(expectedActivities, nil).Once()

		processor := NewIterativeProcessor(msgCtx, func(string) {})

		// Round 1: Profile succeeds
		toolCalls1 := []openai.ToolCall{
			{ID: "call-1", Function: openai.FunctionCall{Name: "get-athlete-profile"}},
		}

		results1, err := service.executeTools(ctx, msgCtx, toolCalls1)
		assert.NoError(t, err)
		assert.Len(t, results1, 1)
		processor.AddToolResults(results1)

		// Round 2: Activities fails
		toolCalls2 := []openai.ToolCall{
			{ID: "call-2", Function: openai.FunctionCall{Name: "get-recent-activities"}},
		}

		results2, err := service.executeToolsWithRecovery(ctx, msgCtx, toolCalls2)
		assert.Error(t, err) // Should fail
		assert.Len(t, results2, 0)

		// Round 3: Activities succeeds (recovery)
		toolCalls3 := []openai.ToolCall{
			{ID: "call-3", Function: openai.FunctionCall{Name: "get-recent-activities"}},
		}

		results3, err := service.executeTools(ctx, msgCtx, toolCalls3)
		assert.NoError(t, err)
		assert.Len(t, results3, 1)
		processor.AddToolResults(results3)

		// Verify analysis can continue with recovered data
		shouldContinue, reason := service.shouldContinueAnalysis(processor, []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-activity-details"}},
		}, false)
		assert.True(t, shouldContinue)
		assert.Equal(t, "continue_analysis", reason)

		// Verify final state shows successful recovery
		assert.Equal(t, 2, processor.CurrentRound) // Only successful rounds
		assert.Equal(t, 2, processor.GetTotalToolCalls())

		service.mockStrava.AssertExpectations(t)
	})

	t.Run("partial tool failure with graceful degradation", func(t *testing.T) {
		service := setupTestAIService()
		ctx := context.Background()

		user := &models.User{
			ID:          "partial-failure",
			AccessToken: "partial-token",
		}

		msgCtx := &MessageContext{
			UserID: "partial-failure",
			User:   user,
		}

		// Mock scenario where some tools succeed and others fail
		expectedProfile := &StravaAthlete{ID: 777777, Firstname: "Partial"}
		service.mockStrava.On("GetAthleteProfile", "partial-token").Return(expectedProfile, nil)

		expectedActivities := []*StravaActivity{
			{ID: 7001, Name: "Test Run", Type: "Run"},
		}
		service.mockStrava.On("GetActivities", "partial-token", mock.AnythingOfType("ActivityParams")).Return(expectedActivities, nil)

		// Activity details fail
		service.mockStrava.On("GetActivityDetail", "partial-token", int64(7001)).Return((*StravaActivityDetail)(nil), fmt.Errorf("activity not found"))

		// Streams also fail
		service.mockStrava.On("GetActivityStreams", "partial-token", int64(7001), mock.AnythingOfType("[]string"), mock.AnythingOfType("string")).Return((*StravaStreams)(nil), fmt.Errorf("streams not available"))

		processor := NewIterativeProcessor(msgCtx, func(string) {})

		// Execute mixed success/failure round
		mixedToolCalls := []openai.ToolCall{
			{ID: "call-1", Function: openai.FunctionCall{Name: "get-athlete-profile"}},
			{ID: "call-2", Function: openai.FunctionCall{Name: "get-recent-activities"}},
			{ID: "call-3", Function: openai.FunctionCall{Name: "get-activity-details", Arguments: `{"activity_id": 7001}`}},
			{ID: "call-4", Function: openai.FunctionCall{Name: "get-activity-streams", Arguments: `{"activity_id": 7001}`}},
		}

		results, err := service.executeToolsWithRecovery(ctx, msgCtx, mixedToolCalls)

		// Should succeed with partial results
		assert.NoError(t, err)
		assert.Len(t, results, 2) // Only successful tools

		// Verify successful results
		successfulCalls := make(map[string]bool)
		for _, result := range results {
			successfulCalls[result.ToolCallID] = true
			assert.Empty(t, result.Error)
		}

		assert.True(t, successfulCalls["call-1"]) // Profile should succeed
		assert.True(t, successfulCalls["call-2"]) // Activities should succeed
		assert.False(t, successfulCalls["call-3"]) // Details should fail
		assert.False(t, successfulCalls["call-4"]) // Streams should fail

		processor.AddToolResults(results)

		// Verify analysis can continue with available data
		depth := service.assessAnalysisDepth(processor, []openai.ToolCall{})
		assert.Equal(t, 2, depth) // Profile + activities available

		shouldContinue, reason := service.shouldContinueAnalysis(processor, []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "update-athlete-logbook"}},
		}, false)
		assert.True(t, shouldContinue)
		assert.Equal(t, "continue_analysis", reason)

		service.mockStrava.AssertExpectations(t)
	})
}

// TestAIService_ConcurrentAnalysisWorkflows tests concurrent multi-turn analysis
func TestAIService_ConcurrentAnalysisWorkflows(t *testing.T) {
	t.Run("multiple concurrent coaching sessions", func(t *testing.T) {
		service := setupTestAIService()
		ctx := context.Background()

		const numSessions = 10
		var wg sync.WaitGroup

		// Setup multiple concurrent users
		users := make([]*models.User, numSessions)
		for i := 0; i < numSessions; i++ {
			users[i] = &models.User{
				ID:          fmt.Sprintf("concurrent-user-%d", i),
				AccessToken: fmt.Sprintf("concurrent-token-%d", i),
			}

			// Mock data for each user
			profile := &StravaAthlete{
				ID:        int64(800000 + i),
				Firstname: fmt.Sprintf("User%d", i),
				Lastname:  "Concurrent",
			}
			service.mockStrava.On("GetAthleteProfile", fmt.Sprintf("concurrent-token-%d", i)).Return(profile, nil)

			activities := []*StravaActivity{
				{ID: int64(8000 + i), Name: fmt.Sprintf("Run %d", i), Type: "Run"},
			}
			service.mockStrava.On("GetActivities", fmt.Sprintf("concurrent-token-%d", i), mock.AnythingOfType("ActivityParams")).Return(activities, nil)
		}

		results := make([]struct {
			userID    string
			processor *IterativeProcessor
			duration  time.Duration
			success   bool
		}, numSessions)

		start := time.Now()

		// Execute concurrent analysis sessions
		for i := 0; i < numSessions; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				sessionStart := time.Now()
				user := users[index]

				msgCtx := &MessageContext{
					UserID:    user.ID,
					SessionID: fmt.Sprintf("concurrent-session-%d", index),
					Message:   fmt.Sprintf("Analyze my training - session %d", index),
					User:      user,
				}

				processor := NewIterativeProcessor(msgCtx, func(string) {})

				// Execute multi-round analysis
				success := true

				// Round 1: Profile
				toolCalls1 := []openai.ToolCall{
					{ID: fmt.Sprintf("call-%d-1", index), Function: openai.FunctionCall{Name: "get-athlete-profile"}},
				}

				results1, err := service.executeTools(ctx, msgCtx, toolCalls1)
				if err != nil {
					success = false
				} else {
					processor.AddToolResults(results1)
				}

				// Round 2: Activities
				if success {
					toolCalls2 := []openai.ToolCall{
						{ID: fmt.Sprintf("call-%d-2", index), Function: openai.FunctionCall{Name: "get-recent-activities"}},
					}

					results2, err := service.executeTools(ctx, msgCtx, toolCalls2)
					if err != nil {
						success = false
					} else {
						processor.AddToolResults(results2)
					}
				}

				// Verify analysis decisions
				if success {
					shouldContinue, _ := service.shouldContinueAnalysis(processor, []openai.ToolCall{}, false)
					// Should stop since no more tool calls
					success = !shouldContinue
				}

				results[index] = struct {
					userID    string
					processor *IterativeProcessor
					duration  time.Duration
					success   bool
				}{
					userID:    user.ID,
					processor: processor,
					duration:  time.Since(sessionStart),
					success:   success,
				}
			}(i)
		}

		wg.Wait()
		totalDuration := time.Since(start)

		// Verify all sessions completed successfully
		for i, result := range results {
			assert.True(t, result.success, "Session %d should succeed", i)
			assert.NotNil(t, result.processor, "Session %d processor should not be nil", i)
			assert.Equal(t, 2, result.processor.CurrentRound, "Session %d should complete 2 rounds", i)
			assert.Less(t, result.duration, 100*time.Millisecond, "Session %d should complete quickly", i)
		}

		// Verify concurrent execution was efficient
		assert.Less(t, totalDuration, 200*time.Millisecond, "Concurrent sessions should complete efficiently")

		service.mockStrava.AssertExpectations(t)
	})

	t.Run("concurrent progress message generation", func(t *testing.T) {
		service := setupTestAIService()

		const numConcurrent = 20
		const messagesPerGoroutine = 50

		var wg sync.WaitGroup
		allMessages := make([][]string, numConcurrent)

		start := time.Now()

		for i := 0; i < numConcurrent; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				processor := NewIterativeProcessor(&MessageContext{}, func(string) {})
				allMessages[index] = make([]string, messagesPerGoroutine)

				toolCalls := []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "get-recent-activities"}},
				}

				for j := 0; j < messagesPerGoroutine; j++ {
					processor.CurrentRound = j % 5
					allMessages[index][j] = service.getCoachingProgressMessage(processor, toolCalls)
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		// Should handle high concurrent load efficiently
		assert.Less(t, duration, 100*time.Millisecond, "Concurrent message generation should be fast")

		// Verify all messages are valid
		totalMessages := 0
		for i, messages := range allMessages {
			for j, message := range messages {
				assert.NotEmpty(t, message, "Message [%d][%d] should not be empty", i, j)
				assert.NotContains(t, message, "API", "Message [%d][%d] should not contain API", i, j)
				totalMessages++
			}
		}

		assert.Equal(t, numConcurrent*messagesPerGoroutine, totalMessages)
	})
}

// TestAIService_LongRunningAnalysisWorkflows tests extended analysis sessions
func TestAIService_LongRunningAnalysisWorkflows(t *testing.T) {
	t.Run("extended multi-round analysis session", func(t *testing.T) {
		service := setupTestAIService()
		ctx := context.Background()

		user := &models.User{
			ID:          "long-analysis",
			AccessToken: "long-token",
		}

		msgCtx := &MessageContext{
			UserID:    "long-analysis",
			SessionID: "long-session",
			Message:   "Provide comprehensive analysis of my entire training history",
			User:      user,
		}

		// Mock extensive data for long analysis
		profile := &StravaAthlete{ID: 999999, Firstname: "Long", Lastname: "Analysis"}
		service.mockStrava.On("GetAthleteProfile", "long-token").Return(profile, nil)

		// Mock many activities
		activities := make([]*StravaActivity, 50)
		for i := 0; i < 50; i++ {
			activities[i] = &StravaActivity{
				ID:   int64(9000 + i),
				Name: fmt.Sprintf("Activity %d", i),
				Type: "Run",
			}
		}
		service.mockStrava.On("GetActivities", "long-token", mock.AnythingOfType("ActivityParams")).Return(activities, nil)

		// Mock details for key activities (simulate selective deep analysis)
		keyActivities := []int{0, 10, 20, 30, 40} // Every 10th activity
		for _, idx := range keyActivities {
			detail := &StravaActivityDetail{
				StravaActivity: *activities[idx],
				Description:    fmt.Sprintf("Key activity %d", idx),
			}
			service.mockStrava.On("GetActivityDetail", "long-token", int64(9000+idx)).Return(detail, nil)
		}

		// Mock logbook update
		logbook := &models.AthleteLogbook{
			UserID:  "long-analysis",
			Content: "Comprehensive analysis completed",
		}
		service.mockLogbook.On("UpdateLogbook", ctx, "long-analysis", mock.AnythingOfType("string")).Return(logbook, nil)

		processor := NewIterativeProcessor(msgCtx, func(string) {})

		start := time.Now()

		// Execute extended analysis workflow
		analysisSteps := []struct {
			name      string
			toolCalls []openai.ToolCall
		}{
			{
				name: "Profile analysis",
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "get-athlete-profile"}},
				},
			},
			{
				name: "Activity overview",
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "get-recent-activities", Arguments: `{"per_page": 50}`}},
				},
			},
			{
				name: "Key activity analysis",
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "get-activity-details", Arguments: `{"activity_id": 9000}`}},
					{Function: openai.FunctionCall{Name: "get-activity-details", Arguments: `{"activity_id": 9010}`}},
					{Function: openai.FunctionCall{Name: "get-activity-details", Arguments: `{"activity_id": 9020}`}},
				},
			},
			{
				name: "Additional key activities",
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "get-activity-details", Arguments: `{"activity_id": 9030}`}},
					{Function: openai.FunctionCall{Name: "get-activity-details", Arguments: `{"activity_id": 9040}`}},
				},
			},
			{
				name: "Comprehensive logbook update",
				toolCalls: []openai.ToolCall{
					{Function: openai.FunctionCall{Name: "update-athlete-logbook"}},
				},
			},
		}

		for i, step := range analysisSteps {
			t.Logf("Executing extended analysis step %d: %s", i+1, step.name)

			// Verify we can continue (unless at max rounds)
			if processor.CurrentRound < processor.MaxRounds {
				shouldContinue, _ := service.shouldContinueAnalysis(processor, step.toolCalls, false)
				assert.True(t, shouldContinue, "Should continue for step %d", i+1)
			}

			// Execute step
			results, err := service.executeTools(ctx, msgCtx, step.toolCalls)
			assert.NoError(t, err, "Step %d should succeed", i+1)
			processor.AddToolResults(results)

			// Verify progress
			assert.Equal(t, i+1, processor.CurrentRound, "Round should match step")
		}

		duration := time.Since(start)

		// Verify extended analysis completed efficiently
		assert.Less(t, duration, 200*time.Millisecond, "Extended analysis should complete efficiently")
		assert.Equal(t, 5, processor.CurrentRound)
		assert.Equal(t, 8, processor.GetTotalToolCalls()) // 1+1+3+2+1

		// Verify final analysis state
		finalDepth := service.assessAnalysisDepth(processor, []openai.ToolCall{})
		assert.Equal(t, 3, finalDepth) // Profile + activities + details

		service.mockStrava.AssertExpectations(t)
		service.mockLogbook.AssertExpectations(t)
	})

	t.Run("analysis with maximum rounds reached", func(t *testing.T) {
		service := setupTestAIService()

		processor := &IterativeProcessor{
			MaxRounds:    3, // Low limit for testing
			CurrentRound: 0,
		}

		// Simulate reaching maximum rounds
		toolCalls := []openai.ToolCall{
			{Function: openai.FunctionCall{Name: "get-athlete-profile"}},
		}

		// Should continue for rounds 0, 1, 2
		for round := 0; round < 3; round++ {
			shouldContinue, reason := service.shouldContinueAnalysis(processor, toolCalls, false)
			if round < 3 {
				assert.True(t, shouldContinue, "Should continue for round %d", round)
				assert.Equal(t, "continue_analysis", reason)
			}
			processor.CurrentRound++
		}

		// Should stop at round 3
		shouldContinue, reason := service.shouldContinueAnalysis(processor, toolCalls, false)
		assert.False(t, shouldContinue)
		assert.Equal(t, "max_rounds", reason)

		// Verify final response acknowledges completion
		finalResponse := service.generateFinalResponse(processor, reason, false)
		assert.NotEmpty(t, finalResponse)
		assert.NotContains(t, finalResponse, "API")
		// Should indicate comprehensive analysis was performed
		assert.Contains(t, finalResponse, "analysis")
	})
}