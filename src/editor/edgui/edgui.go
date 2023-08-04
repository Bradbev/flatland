// edgui provides some helper functions for imgui

package edgui

import (
	"fmt"

	"github.com/inkyblackness/imgui-go/v4"
)

func Text(format string, args ...any) {
	t := fmt.Sprintf(format, args...)
	imgui.Text(t)
}
