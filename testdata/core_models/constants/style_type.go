package constants

type StyleType int

const (
	STYLE_TYPE_MIX_IT_UP StyleType = iota + 1
	STYLE_TYPE_IMAGINATIVE
	STYLE_TYPE_FUTURISTIC
	STYLE_TYPE_RETRO
	STYLE_TYPE_PHOTOREALISTIC
	STYLE_TYPE_ABSTRACT
	STYLE_TYPE_MINIMALIST
	STYLE_TYPE_POP_ART
	STYLE_TYPE_3D_ANIMATED
	STYLE_TYPE_ANIME
	STYLE_TYPE_SYNTHWAVE
	STYLE_TYPE_WATERCOLOR
	STYLE_TYPE_IMPRESSIONISM
	STYLE_TYPE_VECTOR_ART
	STYLE_TYPE_STEAMPUNK
	STYLE_TYPE_CUSTOM
)

func (this StyleType) String() string {
	switch this {
	case STYLE_TYPE_MIX_IT_UP:
		return "Mix It Up"
	case STYLE_TYPE_IMAGINATIVE:
		return "Imaginative"
	case STYLE_TYPE_FUTURISTIC:
		return "Futuristic"
	case STYLE_TYPE_RETRO:
		return "Retro"
	case STYLE_TYPE_PHOTOREALISTIC:
		return "Photorealistic"
	case STYLE_TYPE_ABSTRACT:
		return "Abstract"
	case STYLE_TYPE_MINIMALIST:
		return "Minimalist"
	case STYLE_TYPE_POP_ART:
		return "Pop Art"
	case STYLE_TYPE_3D_ANIMATED:
		return "3D Animated"
	case STYLE_TYPE_ANIME:
		return "Anime"
	case STYLE_TYPE_SYNTHWAVE:
		return "Synthwave"
	case STYLE_TYPE_WATERCOLOR:
		return "Watercolor"
	case STYLE_TYPE_IMPRESSIONISM:
		return "Impressionism"
	case STYLE_TYPE_VECTOR_ART:
		return "Vector Art"
	case STYLE_TYPE_STEAMPUNK:
		return "Steampunk"
	case STYLE_TYPE_CUSTOM:
		return "Custom"
	}
	return ""
}
