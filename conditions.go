package main

import "fmt"

type Operator string

const (
	EQ Operator = "equals"
	EX Operator = "exists"
)

type Condition struct {
	Value    string
	Operator Operator
}

func (C *Condition) Validate(respBody map[string]string, key string) error {
	if C.Operator == EX {
		_, ok := respBody[key]
		if !ok {
			return fmt.Errorf("expected %s to be in the response body but it was not", key)

		}
	}

	if C.Operator == EQ && respBody[key] != C.Value {
		return fmt.Errorf("expected %s to be equal to %s. Got %s instead", key, C.Value, respBody[key])
	}

	return nil
}

func (T *TestCaseResult) ValidateConditions(respBody map[string]string) {

	for key, conditionsForKey := range T.Case.Conditions {

		for _, cond := range conditionsForKey {
			err := cond.Validate(respBody, key)
			if err != nil {
				T.AddErrMsg(fmt.Errorf("error running condition on key %s: %w", key, err).Error())
			}
		}

	}
}
