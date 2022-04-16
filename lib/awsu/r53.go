package awsu

import (
	"fmt"

	"github.com/isan-rivkin/route53-cli/aws_utils"
	"github.com/isan-rivkin/route53-cli/sdk"
)

func NewR53Input(recordInput, awsProfile string, debug, muteLogs bool, skipNSVerification, recursiveSearch bool, recursiveMaxDepth int) (sdk.Input, error) {
	return sdk.NewInput(recordInput, awsProfile, debug, muteLogs, skipNSVerification, recursiveSearch, recursiveMaxDepth)
}

// https://github.com/pterm/pterm/
func SearchRoute53(in sdk.Input) (*sdk.ResultOutput, error) {
	result, err := sdk.SearchR53(in)
	fmt.Println("1111")
	if err != nil {
		fmt.Println("ha wa?")
		return nil, err
	}
	fmt.Println("2222222")
	for _, r := range result {
		fmt.Println("33333")
		r.PrintTable(&aws_utils.PrintOptions{
			WebURL: true,
		})
	}
	fmt.Println("4444")
	return sdk.ToSimpleOutput(result), nil
}
