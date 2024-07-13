package subcommands
import (
	"errors"
	"fmt"
	"os"
)

type AddArgs struct {
	DataType string
	DataURIs []string
}

func Add(a AddArgs){
	fmt.Println(fmt.Sprintf("Add is not fully implemented. dataType: %s, dataURIs: %v", a.DataType, a.DataURIs))
	for i, u := range a.DataURIs {
	    d, err := os.Stat(u)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println(fmt.Sprintf("Failed to add file %s at input list index %v: file does not exist", u, i))
			return
	  	}
		fmt.Println(d.Name())
	}
}