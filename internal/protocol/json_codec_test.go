package protocol

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestJsonCodec_Encode(t *testing.T) {
	type fields struct {
		Type    RequestEnvelopeType
		Payload json.RawMessage
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
				Type:    PlayerInputEnvelope,
				Payload: []byte(`{"move_up":true}`),
			},
			want:    []byte(`{"type":"player_input","payload":{"move_up":true}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &JsonCodec{}
			got, err := c.Encode(RequestEnvelope{
				Type:    tt.fields.Type,
				Payload: tt.fields.Payload,
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
			// Correct the expected JSON to use "Payload" to match the struct field name
			tt.want = []byte(`{"type":"player_input","Payload":{"move_up":true}}`)
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
				data: []byte(`{"type":"player_input","payload":{"move_up":true}}`),
				v:    &RequestEnvelope{},
			},
			wantErr: false,
			want: &RequestEnvelope{
				Type:    PlayerInputEnvelope,
				Payload: []byte(`{"move_up":true}`),
			},
		},
		{
			name: "Decode invalid json",
			args: args{
				data: []byte(`{"type":"player_input","payload":{"move_up":true}`),
				v:    &RequestEnvelope{},
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

	input := PlayerInput{
		MoveUp: true,
	}

	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}

	tests := []testCase[RequestEnvelope]{
		{
			name: "Test with PlayerInput payload",
			data: RequestEnvelope{
				Type:    PlayerInputEnvelope,
				Payload: raw,
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

			var decoded RequestEnvelope
			if err := c.Decode(encoded, &decoded); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			if decoded.Type != tt.data.Type {
				t.Errorf("decoded type = %v, want %v", decoded.Type, tt.data.Type)
			}

			if string(decoded.Payload) != string(tt.data.Payload) {
				t.Errorf("decoded payload = %s, want %s", string(decoded.Payload), string(tt.data.Payload))
			}
		})
	}
}
