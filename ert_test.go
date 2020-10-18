package ert

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestERT(t *testing.T) {
	mux := New()
	mux.NewGroup("_why", TryAll())
	touched := []int{}
	mux.Add("_why", func(trace Trace, topic, body string) error {
		touched = append(touched, 1)
		return nil
	})
	require.Nil(t, mux.Validate(), "mux should be valid")

	mux.Add("_why", func(trace Trace, topic, body string) error {
		touched = append(touched, 2)
		return nil
	})
	require.Nil(t, mux.Validate(), "mux should be valid")

	mux.Report("_why", T("super", "yummy"), "new guide", "last time")
	require.Equal(t, []int{1, 2}, touched)
}

func TestERTError(t *testing.T) {
	mux := New()
	require.EqualError(t, mux.Validate(), "at least on group must be defined")

	mux.NewGroup("something")
	require.EqualError(t, mux.Validate(), "group 'something' hasn't any reporter assigend")

	mux = New()
	mux.NewGroup("something")
	mux.NewGroup("something")
	require.EqualError(t, mux.Validate(), "group 'something' already exists")

	mux = New()
	mux.Add("somethingx", func(Trace, string, string) error {
		return nil
	})
	require.EqualError(t, mux.Validate(), "group 'somethingx' doesn't exists")
}

func TestERTAddGroups(t *testing.T) {
	mux := New()
	touched := []int{}
	mux.AddGroups(Group{
		Name:    "_why",
		Options: []GroupOption{TryAll()},
		Reporters: []Reporter{
			func(trace Trace, topic, body string) error {
				touched = append(touched, 1)
				return nil
			},
			func(trace Trace, topic, body string) error {
				touched = append(touched, 2)
				return nil
			},
		},
	})
	require.Nil(t, mux.Validate(), "mux should be valid")

	mux.Report("_why", T("super", "yummy"), "new guide", "last time")
	require.Equal(t, []int{1, 2}, touched)
}
