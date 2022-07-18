package ferstream

import (
	"testing"
	"time"

	"github.com/kumparan/ferstream/pb"
	"github.com/kumparan/go-utils"
	"github.com/kumparan/tapao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNatsEventMessage_WithEvent(t *testing.T) {
	tests := []struct {
		Name          string
		Given         *NatsEvent
		ExpectedError bool
	}{
		{
			Name: "success without tenant id",
			Given: &NatsEvent{
				ID:     111,
				UserID: 432,
			},
			ExpectedError: false,
		},
		{
			Name: "success with tenant id",
			Given: &NatsEvent{
				ID:       111,
				UserID:   432,
				TenantID: 666,
			},
			ExpectedError: false,
		},
		{
			Name: "empty id",
			Given: &NatsEvent{
				UserID: 432,
			},
			ExpectedError: true,
		},
		{
			Name: "empty user",
			Given: &NatsEvent{
				ID: 111,
			},
			ExpectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := NewNatsEventMessage().WithEvent(test.Given)
			if test.ExpectedError {
				assert.Error(t, result.Error)
				assert.Nil(t, result.NatsEvent)
				return
			}
			assert.EqualValues(t, test.Given, result.NatsEvent)
		})
	}
}

func TestNatsEventMessage_WithBody(t *testing.T) {
	body := []string{"test"}
	result := NewNatsEventMessage().WithBody(body)
	assert.NoError(t, result.Error)
	assert.Equal(t, utils.Dump(body), result.Body)
}

func TestNatsEventMessage_WithOldBody(t *testing.T) {
	body := []string{"test"}
	result := NewNatsEventMessage().WithOldBody(body)
	assert.NoError(t, result.Error)
	assert.Equal(t, utils.Dump(body), result.OldBody)
}

func TestNatsEventMessage_WithRequest(t *testing.T) {
	body := &pb.FindByIDRequest{Id: 121}
	result := NewNatsEventMessage().WithRequest(body)
	assert.NoError(t, result.Error)

	var requestResult pb.FindByIDRequest
	err := tapao.Unmarshal(result.Request, &requestResult, tapao.FallbackWith(tapao.Protobuf))
	assert.NoError(t, err)
	assert.Equal(t, body.Id, requestResult.GetId())
}

func TestNatsEventMessage_Build(t *testing.T) {
	event := &NatsEvent{
		ID:     1,
		UserID: 123,
	}

	body := []string{"test"}
	oldBody := []string{"old test"}
	req := &pb.Greeting{
		Id:        23,
		Name:      "Hai Zob",
		CreatedAt: time.Now().String(),
		UpdatedAt: time.Now().String(),
	}

	t.Run("success", func(t *testing.T) {
		message, err := NewNatsEventMessage().
			WithEvent(event).
			WithBody(body).
			WithRequest(req).
			Build()
		assert.NoError(t, err)
		assert.NotNil(t, message)

		var result NatsEventMessage
		err = tapao.Unmarshal(message, &result, tapao.FallbackWith(tapao.JSON))
		assert.NoError(t, err)
		assert.Equal(t, event.ID, result.NatsEvent.ID)
		assert.Equal(t, event.UserID, result.NatsEvent.UserID)
		assert.Equal(t, utils.Dump(body), result.Body)

		var requestResult pb.Greeting
		err = tapao.Unmarshal(result.Request, &requestResult, tapao.FallbackWith(tapao.Protobuf))
		assert.NoError(t, err)
		assert.Equal(t, req.Id, requestResult.Id)
		assert.Equal(t, req.Name, requestResult.Name)
		assert.Equal(t, req.CreatedAt, requestResult.CreatedAt)
		assert.Equal(t, req.UpdatedAt, requestResult.UpdatedAt)
	})

	t.Run("success with old body", func(t *testing.T) {
		message, err := NewNatsEventMessage().
			WithEvent(event).
			WithBody(body).
			WithOldBody(oldBody).
			WithRequest(req).
			Build()
		assert.NoError(t, err)
		assert.NotNil(t, message)

		var result NatsEventMessage
		err = tapao.Unmarshal(message, &result, tapao.FallbackWith(tapao.JSON))
		assert.NoError(t, err)
		assert.Equal(t, event.ID, result.NatsEvent.ID)
		assert.Equal(t, event.UserID, result.NatsEvent.UserID)
		assert.Equal(t, utils.Dump(body), result.Body)
		assert.Equal(t, utils.Dump(oldBody), result.OldBody)

		var requestResult pb.Greeting
		err = tapao.Unmarshal(result.Request, &requestResult, tapao.FallbackWith(tapao.Protobuf))
		assert.NoError(t, err)
		assert.Equal(t, req.Id, requestResult.Id)
		assert.Equal(t, req.Name, requestResult.Name)
		assert.Equal(t, req.CreatedAt, requestResult.CreatedAt)
		assert.Equal(t, req.UpdatedAt, requestResult.UpdatedAt)
	})

	t.Run("missing nats event", func(t *testing.T) {
		message, err := NewNatsEventMessage().Build()
		assert.Error(t, err)
		assert.Nil(t, message)

		events := []*NatsEvent{
			nil,
			{
				UserID: 131,
			},
			{
				ID: 21,
			},
		}

		for _, e := range events {
			message, err := NewNatsEventMessage().WithEvent(e).Build()
			assert.Error(t, err)
			assert.Nil(t, message)
		}
	})
}

func TestNatsEventMessage_ToJSON(t *testing.T) {
	natsEvent := &NatsEvent{
		ID:     123,
		UserID: 333,
	}
	body := []string{"test"}

	msg := NewNatsEventMessage().
		WithEvent(natsEvent).
		WithBody(body)

	parsed, err := msg.ToJSON()
	require.NoError(t, err)

	expectedRes := "{\"NatsEvent\":{\"id\":123,\"user_id\":333,\"tenant_id\":0,\"subject\":\"\"},\"body\":\"[\\\"test\\\"]\",\"old_body\":\"\",\"request\":null,\"error\":null}"
	assert.Equal(t, expectedRes, parsed)
}

func TestNatsEventMessage_ParseJSON(t *testing.T) {
	json := "{\"NatsEvent\":{\"id\":123,\"user_id\":333,\"tenant_id\":0,\"subject\":\"\"},\"body\":\"[\\\"test\\\"]\",\"old_body\":\"\",\"request\":null,\"error\":null}"

	natsEvent := &NatsEvent{
		ID:     123,
		UserID: 333,
	}
	body := []string{"test"}
	msg := NewNatsEventMessage().WithEvent(natsEvent).WithBody(body)

	parsed, err := ParseJSON(json)
	require.NoError(t, err)

	assert.Equal(t, msg, parsed)
}
