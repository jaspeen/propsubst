package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadProperties(t *testing.T) {
	testProps := `
	test.known.property=knownpropertyvalue
	test.other= otherpropertyvalue
	test.another  =  another value with = and rest
	`

	res := make(map[string]string)
	ReadProperties(testProps, res)

	if v, ok := res["test.known.property"]; !ok {
		t.Error("test.known.property doesn't exist")
	} else {
		assert.Equal(t, v, "knownpropertyvalue")
	}

	if v, ok := res["test.other"]; !ok {
		t.Error("test.other doesn't exist")
	} else {
		assert.Equal(t, v, "otherpropertyvalue")
	}

	if v, ok := res["test.another"]; !ok {
		t.Error("test.another doesn't exist")
	} else {
		assert.Equal(t, v, "another value with = and rest")
	}
}

func TestSubstitute(t *testing.T) {
	testPropsText := `
	test.known.property=knownpropertyvalue
	test.other= otherpropertyvalue
	test.another  =  another value with = and rest
	`
	testProps := make(map[string]string)
	ReadProperties(testPropsText, testProps)

	testText := `
	This is a known property - ${test.known.property}
	And this should be ignored - \${test.known.property}
	And this is another property -\ \ ${test.other}
	And this is unknown property - ${test.unknown}
	And this is default property - ${test.unknown.but.default:http://defaultPropertyValue:666?param1=val1&param2=val\}2#test}
	`

	foundProps := make(map[string]string)
	res, err := substitute(testProps, testText, false, foundProps)

	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, `
	This is a known property - knownpropertyvalue
	And this should be ignored - \${test.known.property}
	And this is another property -\ \ otherpropertyvalue
	And this is unknown property - ${test.unknown}
	And this is default property - http://defaultPropertyValue:666?param1=val1&param2=val\}2#test
	`, res)

	assert.Equal(t, 3, len(foundProps))
	assert.Equal(t, "knownpropertyvalue", foundProps["${test.known.property}"])
	assert.Equal(t, "otherpropertyvalue", foundProps["${test.other}"])
	assert.Equal(t, "http://defaultPropertyValue:666?param1=val1&param2=val\\}2#test", foundProps["${test.unknown.but.default:http://defaultPropertyValue:666?param1=val1&param2=val\\}2#test}"])
}

func TestSubstituteFailNotFound(t *testing.T) {
	testPropsText := `
	test.known.property=knownpropertyvalue
	test.other= otherpropertyvalue
	test.another  =  another value with = and rest
	`
	testProps := make(map[string]string)
	ReadProperties(testPropsText, testProps)

	testText := `
	This is a known property - ${test.known.property}
	And this should be ignored - \${test.known.property}
	And this is another property -\ \ ${test.other}
	And this is unknown property - ${test.unknown}
	`

	foundProps := make(map[string]string)
	_, err := substitute(testProps, testText, true, foundProps)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "test.unknown")
}
