package ert

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

type ErrorLogger interface {
	LogErrorf(error, string, ...interface{})
}

type Trace struct {
	tt []string
}

func (t Trace) String() string {
	return "/" + strings.Join(t.tt, "/")
}

func T(tt ...string) Trace {
	return Trace{tt: tt}
}

func (t Trace) Add(tt ...string) Trace {
	return Trace{tt: append(t.tt, tt...)}
}

type Reporter func(trace Trace, topic, body string) error

type GroupOption interface {
	apply(*group)
}

type GroupOptionFunc func(*group)

func (fun GroupOptionFunc) apply(g *group) {
	fun(g)
}

func TryAll() GroupOption {
	return GroupOptionFunc(func(g *group) {
		g.tryAll = true
	})
}

type Group struct {
	Name      string
	Options   []GroupOption
	Reporters []Reporter
}

type group struct {
	rr     []Reporter
	tryAll bool
}

type Mux struct {
	logger ErrorLogger
	err    error
	gg     map[string]group
	nop    bool
}

func (mux *Mux) NewGroup(name string, groupOptions ...GroupOption) *Mux {
	if mux.err != nil {
		return mux
	}

	_, ok := mux.gg[name]
	if ok {
		mux.err = fmt.Errorf("group '%s' already exists", name)
		return mux
	}

	if mux.gg == nil {
		mux.gg = map[string]group{}
	}

	g := group{}
	for _, o := range groupOptions {
		o.apply(&g)
	}

	mux.gg[name] = g
	return mux
}

func (mux *Mux) Add(name string, r Reporter) *Mux {
	if mux.err != nil {
		return mux
	}

	g, ok := mux.gg[name]
	if !ok {
		mux.err = fmt.Errorf("group '%s' doesn't exists", name)
		return mux
	}

	g.rr = append(g.rr, r)
	mux.gg[name] = g

	return mux
}

func (mux *Mux) AddGroup(group Group) *Mux {
	mux.NewGroup(group.Name, group.Options...)
	for _, r := range group.Reporters {
		mux.Add(group.Name, r)
	}

	return mux
}

func (mux *Mux) AddGroups(groups ...Group) *Mux {
	for _, g := range groups {
		mux.AddGroup(g)
	}

	return mux
}

func (mux *Mux) Validate() (err error) {
	if mux.err != nil {
		return mux.err
	}

	if mux.gg == nil {
		err = errors.New("at least on group must be defined")
		return
	}

	for n, g := range mux.gg {
		if len(g.rr) == 0 {
			err = fmt.Errorf("group '%s' hasn't any reporter assigend", n)
			return
		}
	}

	return
}

const bodyX = `
bad github.com/f9a/ert/mux.Report call from %s:%d: Given report group '%s' doesn't exists! 

This message is sent to all registered groups and reporters in hope that this message gone reach the corresponding developer-team.

Original-Trace: %s

!! Please try to forward this message to the developer-team !!

Best regards,
Ava None
`

func (mux *Mux) Report(name string, trace Trace, topic, body string) {
	if mux.nop {
		return
	}

	g, ok := mux.gg[name]
	if !ok {
		_, file, line, _ := runtime.Caller(1)
		if mux.logger != nil {
			msg := fmt.Errorf("bad github.com/f9a/ert/mux.Report call from %s:%d: Given report group '%s' doesn't exists", file, line, name)
			mux.logger.LogErrorf(msg, "invalid report")
		}

		// Try to report this to every possible reporter.
		// It is better to send this generic error to the wrong group than if nobody notices it.
		// Don't send orignal message because it could leak sensitive data to the wrong group.
		traceX := T("github.com", "f9a", "ert", "mux.Report")
		topicX := fmt.Sprintf("ERROR: ert: Wrong report group '%s'", name)
		bodyX := fmt.Sprintf(bodyX, file, line, name, trace)
		for _, g := range mux.gg {
			for _, r := range g.rr {
				r(traceX, topicX, bodyX)
			}
		}
		return
	}

	for i, r := range g.rr {
		err := r(trace, topic, body)
		if err != nil {
			fmt.Println("error while reporting")
			if mux.logger != nil {
				mux.logger.LogErrorf(err, "reporting to group '%s' via reporter no. %d failed", name, i)
			}
		} else {
			if !g.tryAll {
				return
			}
		}
	}
}

type Option interface {
	apply(*Mux)
}

type OptionFunc func(*Mux)

func (fun OptionFunc) apply(m *Mux) {
	fun(m)
}

func Logger(logger ErrorLogger) Option {
	return OptionFunc(func(m *Mux) {
		m.logger = logger
	})
}

func New(options ...Option) (mux *Mux) {
	mux = &Mux{}
	for _, o := range options {
		o.apply(mux)
	}

	return
}

// NewNop returns a nop mux.
// Everything is the same but when *Mux#Report is called nothing happens.
func NewNop() (mux *Mux) {
	return &Mux{nop: true}
}
