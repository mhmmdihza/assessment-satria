package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	typeFile string
	output   string
)

const (
	typeFileJSON  = "json"
	typeFileText  = "text"
	extensionLog  = ".log"
	extensionText = ".txt"
	extensionJSON = ".json"
)

type JSONStruct struct {
	DateTime string `json:"date_time"`
	Service  string `json:"service"`
	Message  string `json:"message"`
}

var rootCmd = &cobra.Command{
	Use:   "mytools <logfile>",
	Short: "Assessment Satria",
	Long:  `Assessment Satria for Backend Developer.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("The logfile argument is required")
		}
		if typeFile != "" {
			if typeFile != typeFileJSON && typeFile != typeFileText {
				return errors.New("Invalid input for flag --type\n type can only be inputted with value '" + typeFileJSON + "' or '" + typeFileText + "'")
			}
		}
		return convertLogFile(args[0])
		return nil
	},
}

func convertLogFile(logfile string) error {
	outputExt, fileOutput, err := validateFile(logfile)
	if err != nil {
		return err
	}
	file, err := os.Open(logfile)
	if err != nil {
		return err
	}
	defer func() error {
		if err = file.Close(); err != nil {
			return fmt.Errorf("Error when reading log file , please check the permission or path directory %v", err)
		}
		return nil
	}()

	fo, err := os.OpenFile(fileOutput, os.O_RDWR|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Unable to write file: %v", err)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	//for memory efficient
	buf := make([]byte, 32*1024)
	count := 0
	for {
		n, err := file.Read(buf)
		if n > 0 {
			if outputExt == extensionText {
				fo.Write(buf[:n])
			} else if outputExt == extensionJSON {
				if count == 0 {
					fo.Write([]byte("["))
				}
				lineUnix := strings.Split(string(buf[:n]), "\n")
				if len(lineUnix) != 0 {
					fo.Write([]byte(convertArrToJson(lineUnix)))
				}

			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		count++
	}
	if outputExt == extensionJSON {
		fo.Write([]byte("]"))
	}
	return nil
}

func convertArrToJson(arr []string) string {
	res := ""
	for i, v := range arr {
		jsonStruct := &JSONStruct{}
		s := strings.Split(v, ":")
		if len(s) > 4 {
			for i2, v2 := range s {
				if i2 < 3 {
					jsonStruct.DateTime = jsonStruct.DateTime + v2
				} else if i2 == 3 {
					jsonStruct.Service = v2
				} else if i2 >= 4 {
					jsonStruct.Message = jsonStruct.Message + v2
				}
			}
			b, _ := json.Marshal(jsonStruct)
			res = res + string(b)

		}
		if i < len(arr)-2 {
			res = res + ", "
		}
	}
	return res
}

func validateFile(logfile string) (string, string, error) {
	extension := filepath.Ext(logfile)
	if extension != extensionLog {
		return "", "", errors.New("File must be .log type")
	}

	outputExtension := extensionText

	//no need to add contion where typeFile == "text" ,because its already validationg input only can be json or txt
	if typeFile == typeFileJSON {
		outputExtension = extensionJSON
	}

	if output == "" {
		return outputExtension, logfile[0:len(logfile)-len(extension)] + outputExtension, nil
	}

	extensionTarget := filepath.Ext(output)
	if extensionTarget == extensionJSON {
		outputExtension = extensionJSON
	}
	if extensionTarget != extensionText && extensionTarget != extensionJSON {
		return "", "", errors.New("Invalid Output File Type " + output + " - file type can only be .json or .txt")
	}
	if extensionTarget != outputExtension && typeFile != "" {
		return "", "", errors.New("Invalid Output File Type " + output + " - inconsistent file type for input -t = " + typeFile)
	}

	return outputExtension, output, nil
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&typeFile, "type", "t", "", "output type format ('json'/'text')")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "directory location of output file")
	Execute()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
