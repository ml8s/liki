package qiming

import (
	"fmt"
)

// EvaluateNames independently evaluates candidate given names against the qiming database.
func EvaluateNames(givenNames []string, yongShen string, xiShen, jiShen []string) ([]Evaluation, error) {
	if len(givenNames) == 0 {
		return nil, fmt.Errorf("given_names must contain at least one name")
	}
	yongElem := wuxingFromChinese(yongShen)
	if yongShen != "" && yongElem == 0 {
		return nil, fmt.Errorf("invalid yongshen %q", yongShen)
	}
	xiElems, err := parseWuxingList("xishen", xiShen)
	if err != nil {
		return nil, err
	}
	jiElems, err := parseWuxingList("jishen", jiShen)
	if err != nil {
		return nil, err
	}

	results := make([]Evaluation, 0, len(givenNames))
	for _, givenName := range givenNames {
		results = append(results, evaluateName(givenName, yongElem, xiElems, jiElems))
	}
	return results, nil
}

func evaluateName(givenName string, yongElem Wuxing, xiElems, jiElems []Wuxing) Evaluation {
	evaluation := Evaluation{
		GivenName: givenName,
	}
	givenRunes := []rune(givenName)
	if len(givenRunes) < 1 || len(givenRunes) > 2 {
		evaluation.Errors = append(evaluation.Errors, EvaluationError{Code: "invalid_name_length"})
		return evaluation
	}

	characters := make([]Character, 0, len(givenRunes))
	for _, charRune := range givenRunes {
		character, ok := charByRune[charRune]
		if !ok {
			appendError(&evaluation, "character_not_found", string(charRune))
			continue
		}
		if negativeChars[string(charRune)] {
			appendError(&evaluation, "negative_character_forbidden", string(charRune))
			continue
		}
		characters = append(characters, character)
	}
	if len(evaluation.Errors) != 0 {
		return evaluation
	}

	evaluation.Valid = true
	evaluation.Characters = characters
	phonetic := analyzePhonetic(characters)
	evaluation.Phonetic = &phonetic
	if yongElem != 0 || len(xiElems) != 0 || len(jiElems) != 0 {
		evaluation.Wuxing = &WuxingHit{}
		if yongElem != 0 {
			yong := containsElement(characters, yongElem)
			evaluation.Wuxing.Yong = &yong
		}
		if len(xiElems) != 0 {
			xi := containsAnyElement(characters, xiElems)
			evaluation.Wuxing.Xi = &xi
		}
		if len(jiElems) != 0 {
			ji := containsAnyElement(characters, jiElems)
			evaluation.Wuxing.Ji = &ji
		}
	}
	return evaluation
}

func appendError(evaluation *Evaluation, code, char string) {
	for _, existing := range evaluation.Errors {
		if existing.Code == code && existing.Char == char {
			return
		}
	}
	evaluation.Errors = append(evaluation.Errors, EvaluationError{Code: code, Char: char})
}

func containsElement(chars []Character, elem Wuxing) bool {
	if elem == 0 {
		return false
	}
	for _, char := range chars {
		if char.Element == elem {
			return true
		}
	}
	return false
}

func containsAnyElement(chars []Character, elems []Wuxing) bool {
	for _, elem := range elems {
		if containsElement(chars, elem) {
			return true
		}
	}
	return false
}

func parseWuxingList(field string, values []string) ([]Wuxing, error) {
	if len(values) == 0 {
		return nil, nil
	}
	out := make([]Wuxing, 0, len(values))
	for _, value := range values {
		elem := wuxingFromChinese(value)
		if elem == 0 {
			return nil, fmt.Errorf("invalid %s value %q", field, value)
		}
		out = append(out, elem)
	}
	return out, nil
}
