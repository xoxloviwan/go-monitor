// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package store

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson1e5a3b5fDecodeGithubComXoxloviwanGoMonitorInternalStore(in *jlexer.Lexer, out *MemStorage) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "gauge":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				out.Gauge = make(Gauge)
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v1 float64
					v1 = float64(in.Float64())
					(out.Gauge)[key] = v1
					in.WantComma()
				}
				in.Delim('}')
			}
		case "counter":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				out.Counter = make(Counter)
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v2 int64
					v2 = int64(in.Int64())
					(out.Counter)[key] = v2
					in.WantComma()
				}
				in.Delim('}')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson1e5a3b5fEncodeGithubComXoxloviwanGoMonitorInternalStore(out *jwriter.Writer, in MemStorage) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"gauge\":"
		out.RawString(prefix[1:])
		if in.Gauge == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v3First := true
			for v3Name, v3Value := range in.Gauge {
				if v3First {
					v3First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v3Name))
				out.RawByte(':')
				out.Float64(float64(v3Value))
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"counter\":"
		out.RawString(prefix)
		if in.Counter == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v4First := true
			for v4Name, v4Value := range in.Counter {
				if v4First {
					v4First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v4Name))
				out.RawByte(':')
				out.Int64(int64(v4Value))
			}
			out.RawByte('}')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v MemStorage) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson1e5a3b5fEncodeGithubComXoxloviwanGoMonitorInternalStore(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v MemStorage) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson1e5a3b5fEncodeGithubComXoxloviwanGoMonitorInternalStore(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *MemStorage) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson1e5a3b5fDecodeGithubComXoxloviwanGoMonitorInternalStore(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *MemStorage) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson1e5a3b5fDecodeGithubComXoxloviwanGoMonitorInternalStore(l, v)
}