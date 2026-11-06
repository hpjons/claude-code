package xwayland

import (
	"reflect"
	"testing"
)

func env(vars map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := vars[k]
		return v, ok
	}
}

func TestPlan(t *testing.T) {
	cases := []struct {
		name    string
		env     map[string]string
		sockets []string
		want    map[string]string
	}{
		{
			name: "DISPLAY set, GDK_BACKEND unset",
			env:  map[string]string{"DISPLAY": ":0"},
			want: map[string]string{"GDK_BACKEND": "x11"},
		},
		{
			name: "COSMIC default backend is overridden",
			env:  map[string]string{"DISPLAY": ":1", "GDK_BACKEND": "wayland,x11"},
			want: map[string]string{"GDK_BACKEND": "x11"},
		},
		{
			name:    "no DISPLAY, X socket found",
			sockets: []string{"not-a-socket", "X2"},
			want:    map[string]string{"GDK_BACKEND": "x11", "DISPLAY": ":2"},
		},
		{
			name:    "no X server reachable",
			sockets: []string{"Xfoo"},
			want:    nil,
		},
		{
			name: "user-chosen GDK_BACKEND is respected",
			env:  map[string]string{"DISPLAY": ":0", "GDK_BACKEND": "wayland"},
			want: nil,
		},
		{
			name: "opt-out",
			env:  map[string]string{"DISPLAY": ":0", OptOutEnv: "1"},
			want: nil,
		},
		{
			name: "opt-out set to 0 does not opt out",
			env:  map[string]string{"DISPLAY": ":0", OptOutEnv: "0"},
			want: map[string]string{"GDK_BACKEND": "x11"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := plan(env(c.env), c.sockets)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("plan() = %v, want %v", got, c.want)
			}
		})
	}
}
