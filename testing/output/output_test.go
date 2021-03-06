// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package output

import (
	"testing"

	"github.com/stretchrcom/testify/assert"

	"github.com/soumya92/barista/bar"
	"github.com/soumya92/barista/outputs"
)

func TestAssertions(t *testing.T) {
	a := New(t, outputs.Text("a"))
	a.AssertEqual(bar.TextSegment("a"), "same output")
	assert.True(t, a.Expect("should pass"))
	a.AssertText([]string{"a"}, "text")
	// or another way:
	a.At(0).AssertText("a")
	assert.Equal(t, 1, a.Len(), "has 1 segment")

	a = New(t, outputs.Group(
		outputs.Text("a"),
		outputs.Text("b"),
		outputs.Text("c"),
	))
	a.AssertText([]string{"a", "b", "c"}, "multiple segments")

	a = New(t, outputs.Errorf("something"))
	assert.True(t, a.Expect("should pass"))
	errs := a.AssertError("with error output")
	assert.Equal(t, []string{"something"}, errs, "error descriptions")

	a = New(t, outputs.Group(
		outputs.Errorf("something"),
		outputs.Errorf("other thing"),
	))
	errs = a.AssertError("with multiple error segments")
	assert.Equal(t, []string{"something", "other thing"}, errs,
		"error descriptions with multiple segments")

	a = New(t, outputs.Empty())
	assert.True(t, a.Expect("should pass"))
	a.AssertEmpty("empty output")
	assert.Equal(t, 0, a.Len())
}

func TestAssertionErrors(t *testing.T) {
	var out bar.Output
	assertFail := func(testFunc func(Assertions), args ...interface{}) {
		fakeT := &testing.T{}
		a := New(fakeT, out)
		assert.False(t, fakeT.Failed())
		testFunc(a)
		assert.True(t, fakeT.Failed(), args...)
	}

	expectResult := true // Expect false, so start with true.
	assertFail(func(a Assertions) {
		expectResult = a.Expect()
	}, "Expect with no output")
	assert.False(t, expectResult,
		"Expect returns false after failing test")
	assertFail(func(a Assertions) {
		a.At(0).AssertText("foo")
	}, "At() with no output")
	assertFail(func(a Assertions) {
		a.AssertText([]string{"foo"})
	}, "AssertText with no output")
	assertFail(func(a Assertions) {
		a.AssertEqual(outputs.Text("blah"))
	}, "AssertEqual with no output")
	assertFail(func(a Assertions) {
		a.AssertEmpty()
	}, "AssertEmpty with no output")
	lenResult := -1 // expect 0, so start with non-zero value.
	assertFail(func(a Assertions) {
		lenResult = a.Len()
	}, "Len with no output")
	assert.Equal(t, 0, lenResult, "Len returns 0 with no output")
	errStrings := []string{"-init-"}
	assertFail(func(a Assertions) {
		errStrings = a.AssertError()
	}, "AssertError with no output")
	assert.Empty(t, errStrings, "AssertError returns empty result")

	out = outputs.Empty()

	assertFail(func(a Assertions) {
		a.AssertError()
	}, "AssertError on empty output")
	assertFail(func(a Assertions) {
		a.At(0)
	}, "At(0) on empty output")
	assertFail(func(a Assertions) {
		a.AssertEqual(outputs.Text("test"))
	}, "AssertEqual non-empty with empty")
	assertFail(func(a Assertions) {
		a.AssertText([]string{"something"})
	}, "AssertText non-empty with empty")

	out = outputs.Text("testing")

	assertFail(func(a Assertions) {
		a.AssertError()
	}, "AssertError on non-error output")
	assertFail(func(a Assertions) {
		a.At(1)
	}, "At(n) out of bounds")
	assertFail(func(a Assertions) {
		a.AssertEqual(outputs.Text("not testing"))
	}, "AssertEqual with different output")
	assertFail(func(a Assertions) {
		a.AssertText([]string{"testing", "extra"})
	}, "AssertText with extra segment")

	out = outputs.Group(outputs.Text("a"), outputs.Errorf("b"))
	New(t, out).At(1).AssertError("Error segment mixed with non-error")
	assertFail(func(a Assertions) {
		a.AssertError()
	}, "AssertError with one non-error segment")
}

func TestSegmentAssertions(t *testing.T) {
	a := Segment(t, bar.TextSegment("foo").Urgent(true))
	a.AssertEqual(bar.TextSegment("foo").Urgent(true), "same segment")
	a.AssertText("foo", "text")
	// or another way (not recommended, though):
	assert.Equal(t, bar.TextSegment("foo").Urgent(true), a.Segment())

	a = Segment(t, outputs.Errorf("something").Segments()[0])
	err := a.AssertError("with error output")
	assert.Equal(t, "something", err, "error description")

	// TODO: Fix this?
	assert.Panics(t, func() {
		a = Segment(t, bar.Segment{})
		a.AssertText("")
	}, "uninitialised segment")
}

func TestEmptySegmentAssertions(t *testing.T) {
	s := SegmentAssertions{assert: assert.New(t)}
	s.AssertEqual(bar.TextSegment("doesn't matter"),
		"AssertEqual without segment is nop")
	err := s.AssertError("AssertError without segment is nop")
	assert.Empty(t, err, "AssertError returns empty result")
	s.AssertText("whatever", "AssertText without segment is nop")
	assert.Equal(t, bar.Segment{}, s.Segment(),
		"Segment() returns unintialised without segment")
}

func TestSegmentAssertionErrors(t *testing.T) {
	var segment *bar.Segment
	assertFail := func(testFunc func(SegmentAssertions), args ...interface{}) {
		fakeT := &testing.T{}
		s := SegmentAssertions{segment: segment, assert: assert.New(fakeT)}
		assert.False(t, fakeT.Failed())
		testFunc(s)
		assert.True(t, fakeT.Failed(), args...)
	}

	textSegment := bar.TextSegment("test segment")
	segment = &textSegment

	assertFail(func(s SegmentAssertions) {
		s.AssertError()
	}, "AssertError on non-error output")
	assertFail(func(s SegmentAssertions) {
		s.AssertText("not it")
	}, "AssertText with wrong text")
	assertFail(func(s SegmentAssertions) {
		s.AssertEqual(bar.TextSegment("not testing"))
	}, "AssertEqual with different segment")

	errorSegments := outputs.Errorf("404").Segments()
	segment = &errorSegments[0]
	assertFail(func(s SegmentAssertions) {
		s.AssertText("not it")
	}, "AssertText with wrong text")
	assertFail(func(s SegmentAssertions) {
		s.AssertEqual(bar.TextSegment("404"))
	}, "AssertEqual with different segment")
}
