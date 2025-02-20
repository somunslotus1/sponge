package patch

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/zhufuyi/sponge/cmd/sponge/commands/generate"
	"github.com/zhufuyi/sponge/pkg/gofile"
	"github.com/zhufuyi/sponge/pkg/replacer"
)

// GenTypesPbCommand generate types.proto code
func GenTypesPbCommand() *cobra.Command {
	var (
		moduleName string // go.mod module name
		outPath    string // output directory
		targetFile = "api/types/types.proto"
	)

	cmd := &cobra.Command{
		Use:   "gen-types-pb",
		Short: "Generate types.proto code",
		Long: color.HiBlackString(`generate types.proto code

Examples:
  # generate types.proto code.
  sponge patch gen-types-pb --module-name=yourModuleName

  # generate types.proto code and specify the server directory, Note: code generation will be canceled when the latest generated file already exists.
  sponge patch gen-types-pb --out=./yourServerDir
`),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			mdName, _, _ := getNamesFromOutDir(outPath)
			if mdName != "" {
				moduleName = mdName
			} else if moduleName == "" {
				return fmt.Errorf(`required flag(s) "module-name" not set, use "sponge patch gen-types-pb -h" for help`)
			}

			var isEmpty bool
			if outPath == "" {
				isEmpty = true
			} else {
				isEmpty = false
				if gofile.IsExists(targetFile) {
					fmt.Printf("'%s' already exists, no need to generate it.\n", targetFile)
					return nil
				}
			}

			var err error
			outPath, err = runTypesPbCommand(moduleName, outPath)
			if err != nil {
				return err
			}

			if isEmpty {
				fmt.Printf(`
using help:
  move the folder "api" to your project code folder.

`)
			}
			if gofile.IsWindows() {
				targetFile = "\\" + strings.ReplaceAll(targetFile, "/", "\\")
			} else {
				targetFile = "/" + targetFile
			}
			fmt.Printf("generate \"types-pb\" code successfully, out = %s\n", cutPathPrefix(outPath+targetFile))
			return nil
		},
	}

	cmd.Flags().StringVarP(&moduleName, "module-name", "m", "", "module-name is the name of the module in the 'go.mod' file")
	cmd.Flags().StringVarP(&outPath, "out", "o", "", "output directory, default is ./types-pb_<time>, "+
		"if you specify the directory where the web or microservice generated by sponge, the module-name flag can be ignored")

	return cmd
}

func runTypesPbCommand(moduleName string, outPath string) (string, error) {
	subTplName := "types-pb"
	r := generate.Replacers[generate.TplNameSponge]
	if r == nil {
		return "", errors.New("replacer is nil")
	}

	// setting up template information
	subDirs := []string{"api/types"} // only the specified subdirectory is processed, if empty or no subdirectory is specified, it means all files
	ignoreDirs := []string{}         // specify the directory in the subdirectory where processing is ignored
	ignoreFiles := []string{         // specify the files in the subdirectory to be ignored for processing
		"types.pb.go", "types.pb.validate.go",
	}

	r.SetSubDirsAndFiles(subDirs)
	r.SetIgnoreSubDirs(ignoreDirs...)
	r.SetIgnoreSubFiles(ignoreFiles...)
	fields := addTypePbFields(moduleName)
	r.SetReplacementFields(fields)
	_ = r.SetOutputDir(outPath, subTplName)
	if err := r.SaveFiles(); err != nil {
		return "", err
	}

	return r.GetOutputDir(), nil
}

func addTypePbFields(moduleName string) []replacer.Field {
	var fields []replacer.Field

	fields = append(fields, []replacer.Field{
		{
			Old:             "github.com/zhufuyi/sponge",
			New:             moduleName,
			IsCaseSensitive: false,
		},
	}...)

	return fields
}
