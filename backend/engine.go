package backend

import (
	"encoding/csv"
	"os"
	"sort"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/pterm/pterm"
)

//questions is a slice of questions that are used to compare against the query
var questions = make([]string, 0)

//questionAndAnswers is a map of questions and answers used to lookup the answer for a given question
var questionAndAnswers = make(map[string]string)

//QuestionAndAnswer is a struct that contains a question and answer used for the payload in the API
type QuestionAndAnswer struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

//ReadCSV reads a csv file and returns a slice of a slice of strings
func ReadCSV(fileName string) [][]string {
	file, err := os.Open(fileName)
	if err != nil {
		pterm.Error.Println("Error opening file: ", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	data, err := reader.ReadAll()
	if err != nil {
		pterm.Error.Println("Error reading file: ", err)
	}

	return data

}

//ExtractQuestionsAndAnswers takes a csv file and extracts the questions and answers from it
func ExtractQuestionsAndAnswers(fileName string) {
	data := ReadCSV(fileName)
	for _, line := range data {
		for j, field := range line {
			if j == 0 {
				questionAndAnswers[field] = line[1]
				questions = append(questions, field)
			}
		}
	}
}

//RankedQuerySearch takes a query and returns a ranked list of questions according to similarity (Levenshtein distance) that match the query
func RankedQuerySearch(query string) []string {
	rankedResults := make([]string, 0)
	matches := fuzzy.RankFindFold(query, questions)
	sort.Sort(matches)
	for _, match := range matches {
		rankedResults = append(rankedResults, match.Target)
	}
	return rankedResults
}

//GetQuestionsAndAnswers takes the questions from the query and returns QuestionAndAnswer structs
func GetQuestionsAndAnswers(questions []string) []QuestionAndAnswer {
	results := make([]QuestionAndAnswer, 0)
	for _, question := range questions {
		results = append(results, QuestionAndAnswer{Question: question, Answer: questionAndAnswers[question]})
	}
	return results
}
