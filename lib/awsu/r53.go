package awsu

import (
	"github.com/isan-rivkin/route53-cli/aws_utils"
	"github.com/isan-rivkin/route53-cli/sdk"
)

func NewR53Input(recordInput, awsProfile string, debug, muteLogs bool, skipNSVerification, recursiveSearch bool, recursiveMaxDepth int) (sdk.Input, error) {
	return sdk.NewInput(recordInput, awsProfile, debug, muteLogs, skipNSVerification, recursiveSearch, recursiveMaxDepth)
}

// https://github.com/pterm/pterm/
func SearchRoute53(in sdk.Input) (*sdk.ResultOutput, error) {
	result, err := sdk.SearchR53(in)

	if err != nil {
		return nil, err
	}
	for _, r := range result {
		r.PrintTable(&aws_utils.PrintOptions{
			WebURL: true,
		})
	}
	return sdk.ToSimpleOutput(result), nil
}
