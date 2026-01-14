package address

import (
	"github.com/griffnb/core/lib/sanitize"
)

func (this *Address) ToPublicJSON() any {
	return sanitize.SanitizeModel(this, &Structure{})
}
