package utils

import (
	"encoding/json"
	"reflect"
	"testing"

	"survival/internal/core/ports"
)

func TestJsonCodec_Encode(t *testing.T) {
	type fields struct {
		EnvelopeType ports.RequestEnvelopeType
		Payload      json.RawMessage
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "Encode valid request",
			fields: fields{
				EnvelopeType: ports.PlayerInputEnvelope,
				Payload:      []byte(`{"move_up":true}`),
			},
			want:    []byte(`{"envelope_type":"player_input","payload":{"move_up":true}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JsonCodec{}
			got, err := c.Encode(ports.RequestEnvelope{
				EnvelopeType: tt.fields.EnvelopeType,
				Payload:      tt.fields.Payload,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Unmarshal both actual and expected JSON into maps for comparison
			// This makes the test robust against key ordering differences.
			var gotMap, wantMap map[string]interface{}
			if err := json.Unmarshal(got, &gotMap); err != nil {
				t.Fatalf("Failed to unmarshal actual JSON: %v", err)
			}
			// Correct the expected JSON to use "payload" to match the struct field tag
			tt.want = []byte(`{"envelope_type":"player_input","payload":{"move_up":true}}`)
			if err := json.Unmarshal(tt.want, &wantMap); err != nil {
				t.Fatalf("Failed to unmarshal expected JSON: %v", err)
			}

			if !reflect.DeepEqual(gotMap, wantMap) {
				t.Errorf("Encode() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestJsonCodec_Decode(t *testing.T) {
	type args struct {
		data []byte
		v    interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    interface{}
	}{
		{
			name: "Decode valid request",
			args: args{
				data: []byte(`{"envelope_type":"player_input","payload":{"move_up":true}}`),
				v:    &ports.RequestEnvelope{},
			},
			wantErr: false,
			want: &ports.RequestEnvelope{
				EnvelopeType: ports.PlayerInputEnvelope,
				Payload:      []byte(`{"move_up":true}`),
			},
		},
		{
			name: "Decode invalid json",
			args: args{
				data: []byte(`{"envelope_type":"player_input","payload":{"move_up":true}`),
				v:    &ports.RequestEnvelope{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JsonCodec{}
			if tt.name == "Decode invalid json" {
				tt.args.data = []byte(`invalid json`)
			}

			if err := c.Decode(tt.args.data, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !reflect.DeepEqual(tt.args.v, tt.want) {
				t.Errorf("Decode() got = %v, want %v", tt.args.v, tt.want)
			}
		})
	}
}

func TestJsonCodec_EncodeDecode(t *testing.T) {
	type testCase[T any] struct {
		name string
		data T
	}

	input := ports.PlayerInput{
		MoveUp: true,
	}

	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}

	tests := []testCase[ports.RequestEnvelope]{
		{
			name: "Test with PlayerInput payload",
			data: ports.RequestEnvelope{
				EnvelopeType: ports.PlayerInputEnvelope,
				Payload:      raw,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewJsonCodec()

			encoded, err := c.Encode(tt.data)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			var decoded ports.RequestEnvelope
			if err := c.Decode(encoded, &decoded); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			if decoded.EnvelopeType != tt.data.EnvelopeType {
				t.Errorf("decoded envelope_type = %v, want %v", decoded.EnvelopeType, tt.data.EnvelopeType)
			}

			if string(decoded.Payload) != string(tt.data.Payload) {
				t.Errorf("decoded payload = %s, want %s", string(decoded.Payload), string(tt.data.Payload))
			}
		})
	}
}
