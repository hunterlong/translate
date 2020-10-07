package main

import (
	"encoding/csv"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/translate"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

var (
	tr           *translate.Translate
	awsKey       string
	awsSecret    string
	translations []*Text
)

type Text struct {
	Key string
	En  string
	Fr  string
	De  string
	Ru  string
	Sp  string
	Jp  string
	Cn  string
	Ko  string
	It  string
}

func main() {
	InitAWS()

	csvFile := "./languages/data.csv"
	jsDirectory := "./languages"

	if err := TranslateCSV(csvFile); err != nil {
		panic(err)
	}

	langs := []string{"english", "russian", "french", "german", "spanish", "japanese", "chinese", "italian", "korean"}
	for _, v := range langs {
		if err := CreateJS(v, jsDirectory); err != nil {
			panic(err)
		}
	}
}

// TranslateCSV accepts a CSV file for translations, the CSV format should be:
// key,english_value
// hey,Hello
// bye,Goodbye
// the CSV KEY will be used as the JS variable name
func TranslateCSV(filename string) error {
	file, _ := os.Open(filename)

	defer file.Close()
	c := csv.NewReader(file)

	var translations []*Text
	line := 0
	for {
		// Read each record from csv
		record, err := c.Read()
		if err == io.EOF {
			break
		}
		if line == 0 {
			line++
			continue
		}
		if err != nil {
			fmt.Println(err)
			continue
		}
		key := record[0]
		english := record[1]

		translated := TranslateAll(key, english)
		translations = append(translations, translated)

		fmt.Printf("%s | English: %s | French: %s | German: %s | Russian: %s\n", translated.Key, translated.En, translated.Fr, translated.De, translated.Ru)
		line++
	}
	return nil
}

// Translate accepts the english string, and the translated language type
// to return the string value of that requested language
func Translate(val, language string) string {
	input := &translate.TextInput{
		SourceLanguageCode: aws.String("en"),
		TargetLanguageCode: aws.String(language),
		Text:               aws.String(val),
	}
	req, out := tr.TextRequest(input)
	if err := req.Send(); err != nil {
		panic(req.Error)
	}
	return *out.TranslatedText
}

// TranslateAll accepts a key and english value of a string and returns
// a Text object with all translations
func TranslateAll(key, en string) *Text {
	return &Text{
		Key: key,
		En:  en,
		Fr:  Translate(en, "fr"),
		De:  Translate(en, "de"),
		Ru:  Translate(en, "ru"),
		Sp:  Translate(en, "es"),
		Jp:  Translate(en, "ja"),
		Cn:  Translate(en, "zh"),
		Ko:  Translate(en, "ko"),
		It:  Translate(en, "it"),
	}
}

func (t *Text) String(lang string) string {
	switch lang {
	case "english":
		return fmt.Sprintf(`    %s: "%s"`, t.Key, t.En)
	case "russian":
		return fmt.Sprintf(`    %s: "%s"`, t.Key, t.Ru)
	case "spanish":
		return fmt.Sprintf(`    %s: "%s"`, t.Key, t.Sp)
	case "german":
		return fmt.Sprintf(`    %s: "%s"`, t.Key, t.De)
	case "french":
		return fmt.Sprintf(`    %s: "%s"`, t.Key, t.Fr)
	case "japanese":
		return fmt.Sprintf(`    %s: "%s"`, t.Key, t.Jp)
	case "chinese":
		return fmt.Sprintf(`    %s: "%s"`, t.Key, t.Cn)
	case "korean":
		return fmt.Sprintf(`    %s: "%s"`, t.Key, t.Ko)
	case "italian":
		return fmt.Sprintf(`    %s: "%s"`, t.Key, t.It)
	default:
		return fmt.Sprintf(`    %s: "%s"`, t.Key, t.En)
	}
}

// CreateJS accepts the JS file name to be created, and a slice of translations to
// create a dedicated JS file for each language.
func CreateJS(name, directory string) error {
	data := "const " + name + " = {\n"

	var allvars []string
	for _, v := range translations {
		allvars = append(allvars, v.String(name))
	}

	data += strings.Join(allvars, ",\n")

	data += "\n}\n\nexport default " + name

	return ioutil.WriteFile(directory+"/"+name+".js", []byte(data), os.ModePerm)
}

// InitAWS sets up the AWS Translate package using your AWS credientials
// via environment variables: AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
func InitAWS() {
	awsKey = os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecret = os.Getenv("AWS_SECRET_ACCESS_KEY")

	creds := credentials.NewStaticCredentials(awsKey, awsSecret, "")
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: creds,
	})
	if err != nil {
		panic(err)
	}
	tr = translate.New(sess)
}
