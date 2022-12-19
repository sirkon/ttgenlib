package predictor

import (
	"fmt"
	"testing"

	"github.com/posener/complete"
)

func TestPackagePath_Predict(t *testing.T) {
	args := complete.Args{
		Last: "",
	}
	p := PackagePath{}
	fmt.Println(p.Predict(args))
}
