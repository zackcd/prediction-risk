package trigger_service

import (
	"errors"
	"prediction-risk/internal/app/contract"
	trigger_domain "prediction-risk/internal/app/risk/trigger/domain"
	trigger_mock "prediction-risk/internal/app/risk/trigger/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetByID(t *testing.T) {
	triggerID := trigger_domain.NewTriggerID()
	tests := []struct {
		name        string
		triggerID   trigger_domain.TriggerID
		mockSetup   func(*trigger_mock.MockTriggerRepository)
		expectError bool
	}{
		{
			name:      "successful retrieval",
			triggerID: triggerID,
			mockSetup: func(repo *trigger_mock.MockTriggerRepository) {
				trigger := &trigger_domain.Trigger{TriggerID: triggerID}
				repo.On("Get", mock.Anything, triggerID).
					Return(trigger, nil)
			},
			expectError: false,
		},
		{
			name:      "trigger not found",
			triggerID: triggerID,
			mockSetup: func(repo *trigger_mock.MockTriggerRepository) {
				repo.On("Get", mock.Anything, triggerID).
					Return(nil, ErrTriggerNotFound)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(trigger_mock.MockTriggerRepository)
			tt.mockSetup(mockRepo)

			service := NewTriggerService(mockRepo)
			trigger, err := service.GetByID(tt.triggerID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, trigger)
				assert.Equal(t, tt.triggerID, trigger.TriggerID)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*trigger_mock.MockTriggerRepository)
		expectError bool
		expectLen   int
	}{
		{
			name: "successful retrieval of multiple triggers",
			mockSetup: func(repo *trigger_mock.MockTriggerRepository) {
				triggers := []*trigger_domain.Trigger{
					{TriggerID: trigger_domain.NewTriggerID()},
					{TriggerID: trigger_domain.NewTriggerID()},
				}
				repo.On("GetAll", mock.Anything).Return(triggers, nil)
			},
			expectError: false,
			expectLen:   2,
		},
		{
			name: "empty list",
			mockSetup: func(repo *trigger_mock.MockTriggerRepository) {
				repo.On("GetAll", mock.Anything).Return([]*trigger_domain.Trigger{}, nil)
			},
			expectError: false,
			expectLen:   0,
		},
		{
			name: "repository error",
			mockSetup: func(repo *trigger_mock.MockTriggerRepository) {
				repo.On("GetAll", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectError: true,
			expectLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(trigger_mock.MockTriggerRepository)
			tt.mockSetup(mockRepo)

			service := NewTriggerService(mockRepo)
			triggers, err := service.Get()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, triggers, tt.expectLen)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCancelTrigger(t *testing.T) {
	triggerID := trigger_domain.NewTriggerID()
	tests := []struct {
		name        string
		triggerID   trigger_domain.TriggerID
		mockSetup   func(*trigger_mock.MockTriggerRepository)
		expectError bool
	}{
		{
			name:      "successful cancellation",
			triggerID: triggerID,
			mockSetup: func(repo *trigger_mock.MockTriggerRepository) {
				activeTrigger := &trigger_domain.Trigger{
					TriggerID: triggerID,
					Status:    trigger_domain.StatusActive,
				}
				cancelledTrigger := &trigger_domain.Trigger{
					TriggerID: triggerID,
					Status:    trigger_domain.StatusCancelled,
				}

				repo.On("Get", mock.Anything, triggerID).Return(activeTrigger, nil).Once()
				repo.On("Persist", mock.Anything, mock.MatchedBy(func(t *trigger_domain.Trigger) bool {
					return t.Status == trigger_domain.StatusCancelled
				})).Return(nil)
				repo.On("Get", mock.Anything, triggerID).Return(cancelledTrigger, nil).Once()
			},
			expectError: false,
		},
		{
			name:      "already cancelled",
			triggerID: triggerID,
			mockSetup: func(repo *trigger_mock.MockTriggerRepository) {
				cancelledTrigger := &trigger_domain.Trigger{
					TriggerID: triggerID,
					Status:    trigger_domain.StatusCancelled,
				}
				repo.On("Get", mock.Anything, triggerID).Return(cancelledTrigger, nil)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(trigger_mock.MockTriggerRepository)
			tt.mockSetup(mockRepo)

			service := NewTriggerService(mockRepo)
			trigger, err := service.CancelTrigger(tt.triggerID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, trigger)
				assert.Equal(t, trigger_domain.StatusCancelled, trigger.Status)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCreateStopTrigger(t *testing.T) {
	tests := []struct {
		name         string
		contract     contract.ContractIdentifier
		triggerPrice contract.ContractPrice
		limitPrice   *contract.ContractPrice
		mockSetup    func(*trigger_mock.MockTriggerRepository)
		expectError  bool
	}{
		{
			name: "successful creation",
			contract: contract.ContractIdentifier{
				Ticker: "FOO",
				Side:   contract.SideYes,
			},
			triggerPrice: contract.ContractPrice(50),
			limitPrice:   ptr(contract.ContractPrice(45)),
			mockSetup: func(repo *trigger_mock.MockTriggerRepository) {
				repo.On("Persist", mock.Anything, mock.MatchedBy(func(t *trigger_domain.Trigger) bool {
					return t.TriggerType == trigger_domain.TriggerTypeStop
				})).Return(nil)
				repo.On("Get", mock.Anything, mock.AnythingOfType("trigger_domain.TriggerID")).
					Return(&trigger_domain.Trigger{
						TriggerType: trigger_domain.TriggerTypeStop,
						Status:      trigger_domain.StatusActive,
					}, nil)
			},
			expectError: false,
		},
		{
			name: "invalid trigger price",
			contract: contract.ContractIdentifier{
				Ticker: "FOO",
				Side:   contract.SideYes,
			},
			triggerPrice: -1.0,
			limitPrice:   nil,
			mockSetup:    func(repo *trigger_mock.MockTriggerRepository) {},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(trigger_mock.MockTriggerRepository)
			tt.mockSetup(mockRepo)

			service := NewTriggerService(mockRepo)
			trigger, err := service.CreateStopTrigger(tt.contract, tt.triggerPrice, tt.limitPrice)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, trigger)
				assert.Equal(t, trigger_domain.TriggerTypeStop, trigger.TriggerType)
				assert.Equal(t, trigger_domain.StatusActive, trigger.Status)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateStopTrigger(t *testing.T) {
	triggerID := trigger_domain.NewTriggerID()
	tests := []struct {
		name         string
		triggerID    trigger_domain.TriggerID
		triggerPrice *contract.ContractPrice
		limitPrice   *contract.ContractPrice
		mockSetup    func(*trigger_mock.MockTriggerRepository)
		expectError  bool
	}{
		{
			name:         "successful update",
			triggerID:    triggerID,
			triggerPrice: ptr(contract.ContractPrice(50)),
			limitPrice:   ptr(contract.ContractPrice(49)),
			mockSetup: func(repo *trigger_mock.MockTriggerRepository) {
				existingTrigger := &trigger_domain.Trigger{
					TriggerID:   triggerID,
					TriggerType: trigger_domain.TriggerTypeStop,
					Status:      trigger_domain.StatusActive,
					Condition: trigger_domain.TriggerCondition{
						Price: &trigger_domain.PriceRule{
							Threshold: contract.ContractPrice(50),
							Direction: trigger_domain.Below,
						},
					},
					Actions: []trigger_domain.TriggerAction{
						{
							LimitPrice: ptr(contract.ContractPrice(48)),
							Side:       trigger_domain.Sell,
						},
					},
				}

				repo.On("Get", mock.Anything, triggerID).Return(existingTrigger, nil).Once()
				repo.On("Persist", mock.Anything, mock.MatchedBy(func(t *trigger_domain.Trigger) bool {
					return t.TriggerType == trigger_domain.TriggerTypeStop
				})).Return(nil)
				repo.On("Get", mock.Anything, triggerID).Return(existingTrigger, nil).Once()
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(trigger_mock.MockTriggerRepository)
			tt.mockSetup(mockRepo)

			service := NewTriggerService(mockRepo)
			trigger, err := service.UpdateStopTrigger(tt.triggerID, tt.triggerPrice, tt.limitPrice)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, trigger)
				assert.Equal(t, trigger_domain.TriggerTypeStop, trigger.TriggerType)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateTriggerStatus(t *testing.T) {
	testCases := []struct {
		name          string
		setupMock     func(*trigger_mock.MockTriggerRepository)
		currentStatus trigger_domain.TriggerStatus
		newStatus     trigger_domain.TriggerStatus
		expectedError string
	}{
		{
			name: "successful status update from active to triggered",
			setupMock: func(repo *trigger_mock.MockTriggerRepository) {
				trigger := &trigger_domain.Trigger{Status: trigger_domain.StatusActive}

				repo.On("Get", mock.Anything, mock.Anything).Return(trigger, nil).Once()
				repo.On("Persist", mock.Anything, mock.Anything).Return(nil)
			},
			currentStatus: trigger_domain.StatusActive,
			newStatus:     trigger_domain.StatusTriggered,
		},
		{
			name: "get trigger error",
			setupMock: func(repo *trigger_mock.MockTriggerRepository) {
				repo.On("Get", mock.Anything, mock.Anything).Return(nil, errors.New("db error")).Once()
			},
			currentStatus: trigger_domain.StatusActive,
			newStatus:     trigger_domain.StatusTriggered,
			expectedError: "get trigger: db error",
		},
		{
			name: "cannot transition from terminal status",
			setupMock: func(repo *trigger_mock.MockTriggerRepository) {
				trigger := &trigger_domain.Trigger{Status: trigger_domain.StatusTriggered}
				repo.On("Get", mock.Anything, mock.Anything).Return(trigger, nil).Once()
			},
			currentStatus: trigger_domain.StatusTriggered,
			newStatus:     trigger_domain.StatusCancelled,
			expectedError: "cannot transition from terminal status TRIGGERED",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			repo := new(trigger_mock.MockTriggerRepository)
			tc.setupMock(repo)
			service := NewTriggerService(repo)

			// Execute
			updatedTrigger, err := service.UpdateTriggerStatus(trigger_domain.NewTriggerID(), tc.newStatus)

			// Verify
			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, updatedTrigger)
			} else {
				require.NoError(t, err)
				require.NotNil(t, updatedTrigger)
				assert.Equal(t, tc.newStatus, updatedTrigger.Status)
			}

			// Verify all expectations on mock were met
			repo.AssertExpectations(t)
		})
	}
}

func TestTriggerService_validateStatusTransition(t *testing.T) {
	testCases := []struct {
		name          string
		currentStatus trigger_domain.TriggerStatus
		newStatus     trigger_domain.TriggerStatus
		expectedError string
	}{
		{
			name:          "valid transition from active to triggered",
			currentStatus: trigger_domain.StatusActive,
			newStatus:     trigger_domain.StatusTriggered,
		},
		{
			name:          "valid transition from active to cancelled",
			currentStatus: trigger_domain.StatusActive,
			newStatus:     trigger_domain.StatusCancelled,
		},
		{
			name:          "invalid current status",
			currentStatus: "INVALID",
			newStatus:     trigger_domain.StatusTriggered,
			expectedError: "invalid status",
		},
		{
			name:          "invalid new status",
			currentStatus: trigger_domain.StatusActive,
			newStatus:     "INVALID",
			expectedError: "invalid status",
		},
		{
			name:          "transition from terminal status",
			currentStatus: trigger_domain.StatusTriggered,
			newStatus:     trigger_domain.StatusCancelled,
			expectedError: "cannot transition from terminal status TRIGGERED",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewTriggerService(nil) // Repository not needed for validation

			// Execute
			err := service.validateStatusTransition(tc.currentStatus, tc.newStatus)

			// Verify
			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Helper function to create pointers to values
func ptr[T any](v T) *T {
	return &v
}
