package ferstream

import (
	"encoding/json"
	"fmt"
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
			Name: "success with time",
			Given: &NatsEvent{
				ID:     111,
				UserID: 432,
				Time:   time.Now().Format(NatsEventTimeFormat),
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
		{
			Name: "invalid time format",
			Given: &NatsEvent{
				ID:     111,
				UserID: 432,
				Time:   time.Now().Format(time.RFC822),
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
	err := tapao.Unmarshal(result.Request, &requestResult, tapao.With(tapao.Protobuf))
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
		err = tapao.Unmarshal(message, &result, tapao.With(tapao.JSON))
		assert.NoError(t, err)
		assert.Equal(t, event.ID, result.NatsEvent.ID)
		assert.Equal(t, event.UserID, result.NatsEvent.UserID)
		assert.Equal(t, utils.Dump(body), result.Body)

		var requestResult pb.Greeting
		err = tapao.Unmarshal(result.Request, &requestResult, tapao.With(tapao.Protobuf))
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
		err = tapao.Unmarshal(message, &result, tapao.With(tapao.JSON))
		assert.NoError(t, err)
		assert.Equal(t, event.ID, result.NatsEvent.ID)
		assert.Equal(t, event.UserID, result.NatsEvent.UserID)
		assert.Equal(t, utils.Dump(body), result.Body)
		assert.Equal(t, utils.Dump(oldBody), result.OldBody)

		var requestResult pb.Greeting
		err = tapao.Unmarshal(result.Request, &requestResult, tapao.With(tapao.Protobuf))
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

func TestNatsEventMessage_ToJSONString(t *testing.T) {
	now := time.Now().Format(NatsEventTimeFormat)
	t.Run("success", func(t *testing.T) {
		natsEvent := &NatsEvent{
			ID:     123,
			UserID: 333,
			Time:   now,
		}
		body := []string{"test"}

		msg := NewNatsEventMessage().
			WithEvent(natsEvent).
			WithBody(body)

		parsed, err := msg.ToJSONString()
		require.NoError(t, err)
		expectedRes := fmt.Sprintf("{\"NatsEvent\":{\"id\":123,\"id_string\":\"\",\"user_id\":333,\"tenant_id\":0,\"time\":\"%s\",\"subject\":\"\"},\"body\":\"[\\\"test\\\"]\",\"old_body\":\"\",\"request\":null,\"error\":null}", now)
		assert.Equal(t, expectedRes, parsed)
	})

	t.Run("success with IDString", func(t *testing.T) {
		natsEvent := &NatsEvent{
			UserID:   333,
			IDString: "630484ae00f0d71df588a0ab",
			Time:     now,
		}
		body := []string{"test"}

		msg := NewNatsEventMessage().
			WithEvent(natsEvent).
			WithBody(body)

		parsed, err := msg.ToJSONString()
		require.NoError(t, err)
		expectedRes := fmt.Sprintf("{\"NatsEvent\":{\"id\":0,\"id_string\":\"630484ae00f0d71df588a0ab\",\"user_id\":333,\"tenant_id\":0,\"time\":\"%s\",\"subject\":\"\"},\"body\":\"[\\\"test\\\"]\",\"old_body\":\"\",\"request\":null,\"error\":null}", now)
		assert.Equal(t, expectedRes, parsed)
	})
}

func TestNatsEventMessage_ToJSONByte(t *testing.T) {
	natsEvent := &NatsEvent{
		ID:     123,
		UserID: 333,
	}
	body := []string{"test"}

	msg := NewNatsEventMessage().
		WithEvent(natsEvent).
		WithBody(body)

	jsonByte, err := msg.ToJSONByte()
	require.NoError(t, err)

	parsed, err := ParseJSON(string(jsonByte))
	require.NoError(t, err)

	assert.Equal(t, parsed, msg)
}

func TestNatsEventMessage_ParseJSON(t *testing.T) {
	now := time.Now().Format(NatsEventTimeFormat)
	json := fmt.Sprintf("{\"NatsEvent\":{\"id\":123,\"user_id\":333,\"tenant_id\":0,\"time\":\"%s\",\"subject\":\"\"},\"body\":\"[\\\"test\\\"]\",\"old_body\":\"\",\"request\":null,\"error\":null}", now)
	natsEvent := &NatsEvent{
		ID:     123,
		UserID: 333,
		Time:   now,
	}
	body := []string{"test"}
	msg := NewNatsEventMessage().WithEvent(natsEvent).WithBody(body)

	parsed, err := ParseJSON(json)
	require.NoError(t, err)

	assert.Equal(t, msg, parsed)
}

func TestNatsEventAuditLogMessage_Build(t *testing.T) {
	type User struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	oldData := User{
		ID:   int64(123),
		Name: "test name",
	}

	newData := User{
		ID:   int64(123),
		Name: "new test name",
	}

	byteOldData, err := json.Marshal(oldData)
	require.NoError(t, err)
	byteNewData, err := json.Marshal(newData)
	require.NoError(t, err)

	createdAt, err := time.Parse("2006-01-02", "2020-01-29")
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		msg := &NatsEventAuditLogMessage{
			ServiceName:    "test-audit",
			UserID:         123,
			AuditableType:  "user",
			AuditableID:    "123",
			Action:         "update",
			AuditedChanges: string(byteNewData),
			OldData:        string(byteOldData),
			NewData:        string(byteNewData),
			CreatedAt:      createdAt,
			Error:          nil,
		}

		msgByte, err := msg.Build()
		require.NoError(t, err)
		assert.NotNil(t, msgByte)

		var result NatsEventAuditLogMessage
		err = tapao.Unmarshal(msgByte, &result, tapao.With(tapao.JSON))
		assert.NoError(t, err)
		assert.Equal(t, msg.ServiceName, result.ServiceName)
		assert.Equal(t, msg.OldData, result.OldData)
		assert.Equal(t, msg.NewData, result.NewData)
	})
}

func TestNatsEventAuditLogMessage_ToJSONString(t *testing.T) {
	type User struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	oldData := User{
		ID:   int64(123),
		Name: "test name",
	}

	newData := User{
		ID:   int64(123),
		Name: "new test name",
	}

	byteOldData, err := json.Marshal(oldData)
	require.NoError(t, err)
	byteNewData, err := json.Marshal(newData)
	require.NoError(t, err)

	createdAt, err := time.Parse("2006-01-02", "2020-01-29")
	require.NoError(t, err)

	msg := &NatsEventAuditLogMessage{
		ServiceName:    "test-audit",
		UserID:         123,
		AuditableType:  "user",
		AuditableID:    "123",
		Action:         "update",
		AuditedChanges: string(byteNewData),
		OldData:        string(byteOldData),
		NewData:        string(byteNewData),
		CreatedAt:      createdAt,
		Error:          nil,
	}

	parsed, err := msg.ToJSONString()
	require.NoError(t, err)
	expectedRes := "{\"subject\":\"\",\"service_name\":\"test-audit\",\"user_id\":123,\"auditable_type\":\"user\",\"auditable_id\":\"123\",\"action\":\"update\",\"audited_changes\":\"{\\\"id\\\":123,\\\"name\\\":\\\"new test name\\\"}\",\"old_data\":\"{\\\"id\\\":123,\\\"name\\\":\\\"test name\\\"}\",\"new_data\":\"{\\\"id\\\":123,\\\"name\\\":\\\"new test name\\\"}\",\"created_at\":\"2020-01-29T00:00:00Z\",\"error\":null}"
	assert.Equal(t, expectedRes, parsed)
}

func TestNatsEventAuditLogMessage_ToJSONByte(t *testing.T) {
	type User struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	oldData := User{
		ID:   int64(123),
		Name: "test name",
	}

	newData := User{
		ID:   int64(123),
		Name: "new test name",
	}

	byteOldData, err := json.Marshal(oldData)
	require.NoError(t, err)
	byteNewData, err := json.Marshal(newData)
	require.NoError(t, err)

	createdAt, err := time.Parse("2006-01-02", "2020-01-29")
	require.NoError(t, err)

	msg := &NatsEventAuditLogMessage{
		ServiceName:    "test-audit",
		UserID:         123,
		AuditableType:  "user",
		AuditableID:    "123",
		Action:         "update",
		AuditedChanges: string(byteNewData),
		OldData:        string(byteOldData),
		NewData:        string(byteNewData),
		CreatedAt:      createdAt,
		Error:          nil,
	}

	jsonByte, err := msg.ToJSONByte()
	require.NoError(t, err)

	parsed, err := ParseNatsEventAuditLogMessageFromBytes(jsonByte)
	require.NoError(t, err)

	assert.Equal(t, parsed, msg)
}

func TestNatsEventAuditLogMessage_ParseNatsEventAuditLogMessageFromBytes(t *testing.T) {
	payload := "{\"subject\":\"\",\"service_name\":\"test-audit\",\"user_id\":123,\"auditable_type\":\"user\",\"auditable_id\":\"123\",\"action\":\"update\",\"audited_changes\":\"{\\\"id\\\":123,\\\"name\\\":\\\"new test name\\\"}\",\"old_data\":\"{\\\"id\\\":123,\\\"name\\\":\\\"test name\\\"}\",\"new_data\":\"{\\\"id\\\":123,\\\"name\\\":\\\"new test name\\\"}\",\"created_at\":\"2020-01-29T00:00:00Z\",\"error\":null}"

	type User struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	oldData := User{
		ID:   int64(123),
		Name: "test name",
	}

	newData := User{
		ID:   int64(123),
		Name: "new test name",
	}

	byteOldData, err := json.Marshal(oldData)
	require.NoError(t, err)
	byteNewData, err := json.Marshal(newData)
	require.NoError(t, err)

	createdAt, err := time.Parse("2006-01-02", "2020-01-29")
	require.NoError(t, err)

	msg := &NatsEventAuditLogMessage{
		ServiceName:    "test-audit",
		UserID:         123,
		AuditableType:  "user",
		AuditableID:    "123",
		Action:         "update",
		AuditedChanges: string(byteNewData),
		OldData:        string(byteOldData),
		NewData:        string(byteNewData),
		CreatedAt:      createdAt,
		Error:          nil,
	}

	parsed, err := ParseNatsEventAuditLogMessageFromBytes([]byte(payload))
	require.NoError(t, err)

	assert.Equal(t, msg, parsed)
}
